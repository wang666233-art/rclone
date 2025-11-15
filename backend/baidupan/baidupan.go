// Package baidupan provides an interface to Baidu Pan (Baidu Netdisk)
package baidupan

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/rclone/rclone/backend/baidupan/api"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/config/configmap"
	"github.com/rclone/rclone/fs/config/configstruct"
	"github.com/rclone/rclone/fs/config/obscure"
	"github.com/rclone/rclone/fs/fserrors"
	"github.com/rclone/rclone/fs/hash"
	"github.com/rclone/rclone/lib/encoder"
	"github.com/rclone/rclone/lib/oauthutil"
	"github.com/rclone/rclone/lib/pacer"
	"github.com/rclone/rclone/lib/rest"
	xpansdk "open-sdk-go/openxpanapi"
)

const (
	minSleep         = 10 * time.Millisecond
	maxSleep         = 2 * time.Second
	decayConstant    = 2
	defaultChunkSize = 4 * 1024 * 1024 // 4MB - required by Baidu Pan
	maxChunkSize     = 4 * 1024 * 1024 // 4MB - fixed by Baidu Pan
)

// Register with Fs
func init() {
	fs.Register(&fs.RegInfo{
		Name:        "baidupan",
		Description: "Baidu Pan (Baidu Netdisk)",
		NewFs:       NewFs,
		Config:      Config,
		Options: []fs.Option{{
			Name:      "client_id",
			Help:      "Baidu Pan App Key.\n\nLeave blank to use rclone's.",
			Required:  false,
			Sensitive: false,
		}, {
			Name:      "client_secret",
			Help:      "Baidu Pan App Secret.\n\nLeave blank to use rclone's.",
			Required:  false,
			Sensitive: true,
		}, {
			Name:    "auth_flow",
			Help:    "OAuth authorization flow type.",
			Default: "code",
			Examples: []fs.OptionExample{
				{
					Value: "code",
					Help:  "Authorization code flow (recommended, requires browser)",
				}, {
					Value: "device",
					Help:  "Device code flow (for headless environments)",
				},
			},
			Advanced: false,
		}, {
			Name:     "chunk_size",
			Help:     "Upload chunk size. Must be 4MB (fixed by Baidu Pan).",
			Default:  defaultChunkSize,
			Advanced: true,
		}, {
			Name:     config.ConfigEncoding,
			Help:     config.ConfigEncodingHelp,
			Advanced: true,
			Default: (encoder.Base |
				encoder.EncodeBackSlash |
				encoder.EncodeDoubleQuote |
				encoder.EncodeDel |
				encoder.EncodeCtl |
				encoder.EncodeInvalidUtf8),
		}},
	})
}

// Options defines the configuration for this backend
type Options struct {
	ClientID     string               `config:"client_id"`
	ClientSecret string               `config:"client_secret"`
	AuthFlow     string               `config:"auth_flow"`
	ChunkSize    fs.SizeSuffix        `config:"chunk_size"`
	Enc          encoder.MultiEncoder `config:"encoding"`
}

// Fs represents a remote Baidu Pan
type Fs struct {
	name        string
	root        string
	opt         Options
	features    *fs.Features
	srv         *rest.Client
	sdk         *xpansdk.APIClient
	pacer       *fs.Pacer
	dirCache    *dirCache
	tokenSource *oauthutil.TokenSource
	m           configmap.Mapper
}

// Object describes a Baidu Pan file
type Object struct {
	fs      *Fs
	remote  string
	size    int64
	modTime time.Time
	md5     string
	fsid    int64
}

// dirCache caches directory paths to their fsids
type dirCache struct {
	root  string
	cache map[string]int64
}

// Config is called when a new Fs is being configured
func Config(ctx context.Context, name string, m configmap.Mapper, configIn fs.ConfigIn) (*fs.ConfigOut, error) {
	// Get auth flow type
	authFlow, _ := m.Get("auth_flow")
	if authFlow == "" {
		authFlow = "code"
	}

	// Get client ID and secret
	clientID, _ := m.Get("client_id")
	clientSecret, _ := m.Get("client_secret")
	if clientSecret != "" {
		var err error
		clientSecret, err = obscure.Reveal(clientSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to decode client_secret: %w", err)
		}
	}

	oauthConfig := &oauthutil.Options{
		OAuth2Config: &oauthutil.Config{
			AuthURL:      api.AuthURL,
			TokenURL:     api.TokenURL,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"basic", "netdisk"},
			RedirectURL:  oauthutil.RedirectURL,
		},
	}

	// Use device code flow for headless environments
	if authFlow == "device" {
		return configureDeviceFlow(ctx, name, m, configIn, oauthConfig)
	}

	// Default to authorization code flow
	return oauthutil.ConfigOut("", oauthConfig)
}

// configureDeviceFlow handles device code flow for headless environments
func configureDeviceFlow(ctx context.Context, name string, m configmap.Mapper, configIn fs.ConfigIn, oauthConfig *oauthutil.Options) (*fs.ConfigOut, error) {
	// For device code flow, we need to use a different approach
	// Since Baidu Pan's device code flow is non-standard, we'll guide users through it

	switch configIn.State {
	case "":
		// Start device authorization
		return fs.ConfigGoto("device_start")

	case "device_start":
		return fs.ConfigInput("device_info", "config_device_info",
			"Baidu Pan Device Authorization\n\n"+
				"Please follow these steps:\n"+
				"1. Get your Client ID and Client Secret from Baidu Pan Open Platform\n"+
				"2. Visit the authorization URL that will be provided\n"+
				"3. Enter the user code shown\n"+
				"4. Return here and continue\n\n"+
				"Press Enter to continue...")

	case "device_info":
		// In practice, device code flow requires custom implementation
		// For now, fall back to standard OAuth
		fs.Logf(nil, "Device code flow is not yet fully implemented")
		fs.Logf(nil, "Falling back to authorization code flow")
		fs.Logf(nil, "Please make sure you can access a browser")
		return oauthutil.ConfigOut("", oauthConfig)

	default:
		return nil, fmt.Errorf("unknown state %q", configIn.State)
	}
}

// NewFs constructs an Fs from the path, container:path
func NewFs(ctx context.Context, name, root string, m configmap.Mapper) (fs.Fs, error) {
	// Parse config into Options struct
	opt := new(Options)
	err := configstruct.Set(m, opt)
	if err != nil {
		return nil, err
	}

	// Validate chunk size
	if opt.ChunkSize != fs.SizeSuffix(defaultChunkSize) {
		fs.Logf(nil, "Warning: chunk_size must be 4MB for Baidu Pan, using 4MB")
		opt.ChunkSize = fs.SizeSuffix(defaultChunkSize)
	}

	root = strings.Trim(root, "/")

	// Prepare OAuth config
	clientSecret := opt.ClientSecret
	if clientSecret != "" {
		clientSecret = obscure.MustReveal(clientSecret)
	}

	// Get OAuth client
	oAuthClient, ts, err := oauthutil.NewClient(ctx, name, m, &oauthutil.Config{
		AuthURL:      api.AuthURL,
		TokenURL:     api.TokenURL,
		ClientID:     opt.ClientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"basic", "netdisk"},
		RedirectURL:  oauthutil.RedirectURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to configure Baidu Pan OAuth client: %w", err)
	}

	sdkCfg := xpansdk.NewConfiguration()
	sdkCfg.HTTPClient = oAuthClient
	if ua := fs.GetConfig(ctx).UserAgent; ua != "" {
		sdkCfg.UserAgent = ua
	}

	f := &Fs{
		name:        name,
		root:        root,
		opt:         *opt,
		srv:         rest.NewClient(oAuthClient).SetRoot(api.APIBaseURL),
		sdk:         xpansdk.NewAPIClient(sdkCfg),
		pacer:       fs.NewPacer(ctx, pacer.NewDefault(pacer.MinSleep(minSleep), pacer.MaxSleep(maxSleep), pacer.DecayConstant(decayConstant))),
		tokenSource: ts,
		m:           m,
		dirCache: &dirCache{
			root:  root,
			cache: make(map[string]int64),
		},
	}

	f.features = (&fs.Features{
		CaseInsensitive:         false,
		CanHaveEmptyDirectories: true,
		ReadMimeType:            false,
		WriteMimeType:           false,
		Move:                    f.Move,
		Copy:                    f.Copy,
		About:                   f.About,
		PublicLink:              f.PublicLink,
	}).Fill(ctx, f)

	// Check connection by getting user info
	_, err = f.getUserInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return f, nil
}

// Name of the remote (as passed into NewFs)
func (f *Fs) Name() string {
	return f.name
}

// Root of the remote (as passed into NewFs)
func (f *Fs) Root() string {
	return f.root
}

// String converts this Fs to a string
func (f *Fs) String() string {
	return fmt.Sprintf("Baidu Pan root '%s'", f.root)
}

// Precision of the remote
func (f *Fs) Precision() time.Duration {
	return time.Second
}

// Hashes returns the supported hash sets
func (f *Fs) Hashes() hash.Set {
	return hash.Set(hash.MD5)
}

// Features returns the optional features of this Fs
func (f *Fs) Features() *fs.Features {
	return f.features
}

// getAccessToken gets the current access token
func (f *Fs) getAccessToken(ctx context.Context) (string, error) {
	token, err := f.tokenSource.Token()
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

// getUserInfo gets user information
func (f *Fs) getUserInfo(ctx context.Context) (*api.UserInfo, error) {
	accessToken, err := f.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	result, err := callTyped(f, ctx, func() (xpansdk.Uinforesponse, *http.Response, error) {
		return f.sdk.UserinfoApi.Xpannasuinfo(ctx).AccessToken(accessToken).Execute()
	})
	if err != nil {
		return nil, err
	}
	if result.GetErrno() != 0 {
		return nil, &api.Error{
			ErrorCode: int(result.GetErrno()),
			ErrorMsg:  result.GetErrmsg(),
			RequestID: result.GetRequestId(),
		}
	}

	info := &api.UserInfo{
		BaiduName:   result.GetBaiduName(),
		NetdiskName: result.GetNetdiskName(),
		AvatarURL:   result.GetAvatarUrl(),
		VIPType:     int(result.GetVipType()),
		UK:          int64(result.GetUk()),
	}

	return info, nil
}

// getQuotaInfo gets quota information
func (f *Fs) getQuotaInfo(ctx context.Context) (*api.QuotaInfo, error) {
	accessToken, err := f.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	result, err := callTyped(f, ctx, func() (xpansdk.Quotaresponse, *http.Response, error) {
		return f.sdk.UserinfoApi.Apiquota(ctx).
			AccessToken(accessToken).
			Checkfree(1).
			Checkexpire(1).
			Execute()
	})
	if err != nil {
		return nil, err
	}
	if result.GetErrno() != 0 {
		return nil, &api.Error{
			ErrorCode: int(result.GetErrno()),
			ErrorMsg:  fmt.Sprintf("quota error code %d", result.GetErrno()),
		}
	}

	quota := &api.QuotaInfo{
		Total: result.GetTotal(),
		Used:  result.GetUsed(),
		Free:  result.GetFree(),
	}

	return quota, nil
}

// shouldRetry determines if an error is retryable
func (f *Fs) shouldRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if fserrors.ContextError(ctx, &err) {
		return false, err
	}

	if resp != nil && resp.StatusCode == 429 {
		return true, err
	}

	return fserrors.ShouldRetry(err), err
}

func normalizeAPIError(err error) error {
	if err == nil {
		return nil
	}
	var apiErr xpansdk.GenericOpenAPIError
	if errors.As(err, &apiErr) {
		body := apiErr.Body()
		if len(body) > 0 {
			var parsed struct {
				Errno     int         `json:"errno"`
				Errmsg    string      `json:"errmsg"`
				RequestID interface{} `json:"request_id"`
				ErrorCode int         `json:"error_code"`
				ErrorMsg  string      `json:"error_msg"`
			}
			if json.Unmarshal(body, &parsed) == nil {
				code := parsed.Errno
				msg := parsed.Errmsg
				if code == 0 && parsed.ErrorCode != 0 {
					code = parsed.ErrorCode
					msg = parsed.ErrorMsg
				}
				if code != 0 {
					return &api.Error{
						ErrorCode: code,
						ErrorMsg:  msg,
						RequestID: parsed.RequestID,
					}
				}
			}
		}
		if apiErr.Error() != "" {
			return fmt.Errorf("%s", apiErr.Error())
		}
		return errors.New("baidu pan api error")
	}
	return err
}

func callTyped[T any](f *Fs, ctx context.Context, exec func() (T, *http.Response, error)) (T, error) {
	var (
		res  T
		resp *http.Response
		err  error
	)
	err = f.pacer.Call(func() (bool, error) {
		res, resp, err = exec()
		err = normalizeAPIError(err)
		return f.shouldRetry(ctx, resp, err)
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return res, nil
}

func (f *Fs) callJSON(ctx context.Context, exec func() (string, *http.Response, error), out interface{}) error {
	var (
		payload string
		resp    *http.Response
		err     error
	)
	err = f.pacer.Call(func() (bool, error) {
		payload, resp, err = exec()
		err = normalizeAPIError(err)
		return f.shouldRetry(ctx, resp, err)
	})
	if err != nil {
		return err
	}
	if payload == "" {
		payload = "{}"
	}
	return json.Unmarshal([]byte(payload), out)
}

func (f *Fs) callResponseJSON(ctx context.Context, exec func() (*http.Response, error), out interface{}) error {
	return f.callJSON(ctx, func() (string, *http.Response, error) {
		resp, err := exec()
		if resp == nil {
			return "", resp, err
		}
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(data))
		if err == nil && readErr != nil {
			err = readErr
		}
		return string(data), resp, err
	}, out)
}

func (f *Fs) fileManager(ctx context.Context, exec func(accessToken string) (*http.Response, error)) (*api.FileManagerResponse, error) {
	accessToken, err := f.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	var result api.FileManagerResponse
	if err := f.callResponseJSON(ctx, func() (*http.Response, error) {
		return exec(accessToken)
	}, &result); err != nil {
		return nil, err
	}
	if result.ErrorCode != 0 {
		return nil, &result.Error
	}
	return &result, nil
}

// makePath converts a relative path to an absolute Baidu Pan path
func (f *Fs) makePath(remote string) string {
	if f.root == "" {
		return "/" + remote
	}
	return "/" + path.Join(f.root, remote)
}

// dirPath returns the directory path (without trailing slash)
func (f *Fs) dirPath(remote string) string {
	return path.Dir(f.makePath(remote))
}

// List the objects and directories in dir into entries
func (f *Fs) List(ctx context.Context, dir string) (entries fs.DirEntries, err error) {
	accessToken, err := f.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	dirPath := f.makePath(dir)

	var result api.FileListResponse
	err = f.callJSON(ctx, func() (string, *http.Response, error) {
		return f.sdk.FileinfoApi.Xpanfilelist(ctx).
			AccessToken(accessToken).
			Dir(dirPath).
			Folder("0").
			Web("1").
			Execute()
	}, &result)
	if err != nil {
		return nil, err
	}

	if result.ErrorCode != 0 {
		if result.ErrorCode == -9 {
			return nil, fs.ErrorDirNotFound
		}
		return nil, &result.Error
	}

	for _, item := range result.List {
		remote := f.opt.Enc.ToStandardPath(path.Base(item.Path))
		if dir != "" {
			remote = path.Join(dir, remote)
		}

		if item.IsDir() {
			d := fs.NewDir(remote, item.ModTime())
			d.SetID(fmt.Sprintf("%d", item.FsID))
			entries = append(entries, d)
		} else {
			o := &Object{
				fs:      f,
				remote:  remote,
				size:    item.Size,
				modTime: item.ModTime(),
				md5:     item.MD5,
				fsid:    item.FsID,
			}
			entries = append(entries, o)
		}
	}

	return entries, nil
}

// NewObject finds the Object at remote
func (f *Fs) NewObject(ctx context.Context, remote string) (fs.Object, error) {
	return f.newObjectWithInfo(ctx, remote, nil)
}

// newObjectWithInfo creates an Object with optional FileItem info
func (f *Fs) newObjectWithInfo(ctx context.Context, remote string, info *api.FileItem) (fs.Object, error) {
	o := &Object{
		fs:     f,
		remote: remote,
	}

	if info != nil {
		o.setMetaData(info)
	} else {
		err := o.readMetaData(ctx)
		if err != nil {
			return nil, err
		}
	}

	return o, nil
}

// setMetaData sets metadata from a FileItem
func (o *Object) setMetaData(info *api.FileItem) {
	o.size = info.Size
	o.modTime = info.ModTime()
	o.md5 = info.MD5
	o.fsid = info.FsID
}

// readMetaData reads metadata for an object by listing its parent directory
func (o *Object) readMetaData(ctx context.Context) error {
	dir := path.Dir(o.remote)
	if dir == "." {
		dir = ""
	}

	fs.Debugf(o, "readMetaData: remote=%q, dir=%q", o.remote, dir)

	entries, err := o.fs.List(ctx, dir)
	if err != nil {
		fs.Debugf(o, "readMetaData: List failed: %v", err)
		if err == fs.ErrorDirNotFound {
			return fs.ErrorObjectNotFound
		}
		return err
	}

	fs.Debugf(o, "readMetaData: got %d entries", len(entries))

	for _, entry := range entries {
		entryRemote := entry.Remote()
		fs.Debugf(o, "readMetaData: comparing %q == %q", entryRemote, o.remote)
		if entryRemote == o.remote {
			if info, ok := entry.(*Object); ok {
				o.size = info.size
				o.modTime = info.modTime
				o.md5 = info.md5
				o.fsid = info.fsid
				fs.Debugf(o, "readMetaData: matched, size=%d, fsid=%d", o.size, o.fsid)

				// If size is 0, fetch complete metadata using filemetas API
				if o.size == 0 && o.fsid != 0 {
					meta, err := o.getFileMetaWithDlink(ctx)
					if err == nil {
						o.size = meta.Size
						o.md5 = meta.MD5
						o.modTime = time.Unix(meta.ServerMtime, 0)
						fs.Debugf(o, "readMetaData: fetched metadata, size=%d", o.size)
					} else {
						fs.Debugf(o, "readMetaData: failed to fetch metadata: %v", err)
					}
				}

				return nil
			}
			return fs.ErrorIsDir
		}
	}

	fs.Debugf(o, "readMetaData: file not found in list")
	return fs.ErrorObjectNotFound
}

// Put uploads a file
func (f *Fs) Put(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) (fs.Object, error) {
	remote := src.Remote()
	size := src.Size()

	// Create parent directories if needed
	dirPath := path.Dir(remote)
	if dirPath != "" && dirPath != "." {
		err := f.Mkdir(ctx, dirPath)
		if err != nil {
			return nil, err
		}
	}

	// Upload the file
	result, err := f.uploadFile(ctx, in, size, remote, options...)
	if err != nil {
		return nil, err
	}

	// Create the object
	o := &Object{
		fs:      f,
		remote:  remote,
		size:    result.Size,
		modTime: time.Unix(result.ServerMtime, 0),
		md5:     result.MD5,
		fsid:    result.FsID,
	}

	return o, nil
}

// Mkdir makes a directory
func (f *Fs) Mkdir(ctx context.Context, dir string) error {
	accessToken, err := f.getAccessToken(ctx)
	if err != nil {
		return err
	}

	dirPath := f.makePath(dir)

	opts := rest.Opts{
		Method: "POST",
		Path:   api.PathFileManager,
		Parameters: map[string][]string{
			"access_token": {accessToken},
			"method":       {"create"},
		},
		MultipartParams: map[string][]string{
			"path":  {dirPath},
			"isdir": {"1"},
			"rtype": {"0"}, // 0: fail if exists, 1: auto rename, 2: overwrite
		},
	}

	var result api.CreateDirResponse
	var resp *http.Response
	err = f.pacer.Call(func() (bool, error) {
		resp, err = f.srv.CallJSON(ctx, &opts, nil, &result)
		return f.shouldRetry(ctx, resp, err)
	})
	if err != nil {
		return err
	}

	if result.ErrorCode != 0 {
		// -8 means directory already exists, which is not an error for Mkdir
		if result.ErrorCode == -8 {
			return nil
		}
		return &result.Error
	}

	return nil
}

// Rmdir removes a directory
func (f *Fs) Rmdir(ctx context.Context, dir string) error {
	dirPath := f.makePath(dir)

	// First check if directory is empty
	entries, err := f.List(ctx, dir)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fs.ErrorDirectoryNotEmpty
	}

	_, err = f.fileManager(ctx, func(accessToken string) (*http.Response, error) {
		return f.sdk.FilemanagerApi.Filemanagerdelete(ctx).
			AccessToken(accessToken).
			Async(0).
			Filelist(fmt.Sprintf(`["%s"]`, dirPath)).
			Execute()
	})
	return err
}

// ------------------------------------------------------------
// Object interface implementation
// ------------------------------------------------------------

// Fs returns the parent Fs
func (o *Object) Fs() fs.Info {
	return o.fs
}

// Return a string version
func (o *Object) String() string {
	if o == nil {
		return "<nil>"
	}
	return o.remote
}

// Remote returns the remote path
func (o *Object) Remote() string {
	return o.remote
}

// ModTime returns the modification time of the object
func (o *Object) ModTime(ctx context.Context) time.Time {
	return o.modTime
}

// Size returns the size of the object
func (o *Object) Size() int64 {
	return o.size
}

// Hash returns the MD5 of the object
func (o *Object) Hash(ctx context.Context, t hash.Type) (string, error) {
	if t != hash.MD5 {
		return "", hash.ErrUnsupported
	}
	return o.md5, nil
}

// Storable returns whether this object is storable
func (o *Object) Storable() bool {
	return true
}

// SetModTime sets the modification time of the object
func (o *Object) SetModTime(ctx context.Context, modTime time.Time) error {
	// Baidu Pan doesn't support setting modification time
	return fs.ErrorCantSetModTime
}

// Open opens the file for read
func (o *Object) Open(ctx context.Context, options ...fs.OpenOption) (io.ReadCloser, error) {
	fs.Debugf(o, "Opening file for read, fsid=%d, size=%d", o.fsid, o.size)

	// Get download link
	dlink, err := o.getDownloadLink(ctx)
	if err != nil {
		fs.Errorf(o, "Failed to get download link: %v", err)
		return nil, err
	}
	fs.Debugf(o, "Got download link: %s", dlink)

	// Get access token to append to download link
	accessToken, err := o.fs.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// Append access_token to download link
	dlinkWithToken := dlink + "&access_token=" + accessToken
	fs.Debugf(o, "Download URL: %s", dlinkWithToken)

	// Prepare download request with proper User-Agent
	// The HTTP client will automatically follow redirects to the real download URL
	opts := rest.Opts{
		Method:  "GET",
		RootURL: dlinkWithToken,
		Options: options,
		ExtraHeaders: map[string]string{
			"User-Agent": "pan.baidu.com",
		},
	}

	// Make the request
	var resp *http.Response
	err = o.fs.pacer.Call(func() (bool, error) {
		resp, err = o.fs.srv.Call(ctx, &opts)
		return o.fs.shouldRetry(ctx, resp, err)
	})
	if err != nil {
		fs.Errorf(o, "Download request failed: %v", err)
		return nil, err
	}

	fs.Debugf(o, "Download response status: %d, content-length: %d", resp.StatusCode, resp.ContentLength)

	return resp.Body, nil
}

// getFileMetaWithDlink gets file metadata including download link
func (o *Object) getFileMetaWithDlink(ctx context.Context) (*api.FileMeta, error) {
	accessToken, err := o.fs.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var result api.FileMetasResponse
	err = o.fs.callJSON(ctx, func() (string, *http.Response, error) {
		return o.fs.sdk.MultimediafileApi.Xpanmultimediafilemetas(ctx).
			AccessToken(accessToken).
			Fsids(fmt.Sprintf("[%d]", o.fsid)).
			Dlink("1").
			Execute()
	}, &result)
	if err != nil {
		return nil, err
	}

	if result.ErrorCode != 0 {
		return nil, &result.Error
	}

	if len(result.List) == 0 {
		return nil, fs.ErrorObjectNotFound
	}

	return &result.List[0], nil
}

// getDownloadLink gets the download link for the object
func (o *Object) getDownloadLink(ctx context.Context) (string, error) {
	meta, err := o.getFileMetaWithDlink(ctx)
	if err != nil {
		return "", err
	}

	// Update object metadata if not set
	if o.size == 0 {
		o.size = meta.Size
		o.md5 = meta.MD5
		o.modTime = time.Unix(meta.ServerMtime, 0)
	}

	dlink := meta.Dlink
	if dlink == "" {
		return "", fmt.Errorf("no download link available")
	}

	return dlink, nil
}

// Update updates the object with new data
func (o *Object) Update(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) error {
	// Upload new file
	newObj, err := o.fs.Put(ctx, in, src, options...)
	if err != nil {
		return err
	}

	// Update metadata
	if newObject, ok := newObj.(*Object); ok {
		o.size = newObject.size
		o.modTime = newObject.modTime
		o.md5 = newObject.md5
		o.fsid = newObject.fsid
	}

	return nil
}

// Remove deletes the object
func (o *Object) Remove(ctx context.Context) error {
	filePath := o.fs.makePath(o.remote)
	_, err := o.fs.fileManager(ctx, func(accessToken string) (*http.Response, error) {
		return o.fs.sdk.FilemanagerApi.Filemanagerdelete(ctx).
			AccessToken(accessToken).
			Async(0).
			Filelist(fmt.Sprintf(`["%s"]`, filePath)).
			Execute()
	})
	return err
}

// ID returns the ID of the object
func (o *Object) ID() string {
	return fmt.Sprintf("%d", o.fsid)
}

// Move src to this remote using server-side move operations
func (f *Fs) Move(ctx context.Context, src fs.Object, remote string) (fs.Object, error) {
	srcObj, ok := src.(*Object)
	if !ok {
		fs.Debugf(src, "Can't move - not same remote type")
		return nil, fs.ErrorCantMove
	}

	srcPath := srcObj.fs.makePath(srcObj.remote)
	dstPath := f.makePath(remote)

	// Create parent directory if needed
	dirPath := path.Dir(remote)
	if dirPath != "" && dirPath != "." {
		err := f.Mkdir(ctx, dirPath)
		if err != nil {
			return nil, err
		}
	}

	filelist := fmt.Sprintf(`[{"path":"%s","dest":"%s","newname":"%s"}]`,
		srcPath, path.Dir(dstPath), path.Base(dstPath))
	if _, err := f.fileManager(ctx, func(accessToken string) (*http.Response, error) {
		return f.sdk.FilemanagerApi.Filemanagermove(ctx).
			AccessToken(accessToken).
			Async(0).
			Filelist(filelist).
			Execute()
	}); err != nil {
		return nil, err
	}

	return f.NewObject(ctx, remote)
}

// Copy src to this remote using server-side copy operations
func (f *Fs) Copy(ctx context.Context, src fs.Object, remote string) (fs.Object, error) {
	srcObj, ok := src.(*Object)
	if !ok {
		fs.Debugf(src, "Can't copy - not same remote type")
		return nil, fs.ErrorCantCopy
	}

	srcPath := srcObj.fs.makePath(srcObj.remote)
	dstPath := f.makePath(remote)

	// Create parent directory if needed
	dirPath := path.Dir(remote)
	if dirPath != "" && dirPath != "." {
		err := f.Mkdir(ctx, dirPath)
		if err != nil {
			return nil, err
		}
	}

	filelist := fmt.Sprintf(`[{"path":"%s","dest":"%s","newname":"%s"}]`,
		srcPath, path.Dir(dstPath), path.Base(dstPath))
	if _, err := f.fileManager(ctx, func(accessToken string) (*http.Response, error) {
		return f.sdk.FilemanagerApi.Filemanagercopy(ctx).
			AccessToken(accessToken).
			Async(0).
			Filelist(filelist).
			Execute()
	}); err != nil {
		return nil, err
	}

	return f.NewObject(ctx, remote)
}

// About gets quota information
func (f *Fs) About(ctx context.Context) (*fs.Usage, error) {
	quota, err := f.getQuotaInfo(ctx)
	if err != nil {
		return nil, err
	}

	usage := &fs.Usage{
		Total: fs.NewUsageValue(quota.Total),
		Used:  fs.NewUsageValue(quota.Used),
		Free:  fs.NewUsageValue(quota.Free),
	}

	return usage, nil
}

// PublicLink generates a public link for a file or directory
func (f *Fs) PublicLink(ctx context.Context, remote string, expire fs.Duration, unlink bool) (string, error) {
	// Get the object to find its fsid
	obj, err := f.NewObject(ctx, remote)
	if err != nil {
		return "", err
	}

	fsid := obj.(*Object).fsid
	filePath := f.makePath(remote)

	if unlink {
		// Cancel share - not implemented yet
		return "", fs.ErrorNotImplemented
	}

	// Create share link
	opts := rest.Opts{
		Method: "POST",
		Path:   "/rest/2.0/xpan/share",
		Parameters: map[string][]string{
			"method": {"set"},
		},
		MultipartParams: map[string][]string{
			"schannel":     {"0"},
			"channel_list": {"[]"},
			"period":       {"0"}, // 0: permanent, 1: 1 day, 7: 7 days
			"fid_list":     {fmt.Sprintf("[%d]", fsid)},
			"path_list":    {fmt.Sprintf(`["%s"]`, filePath)},
		},
	}

	var result struct {
		api.Error
		Link string `json:"link"`
	}

	var resp *http.Response
	err = f.pacer.Call(func() (bool, error) {
		resp, err = f.srv.CallJSON(ctx, &opts, nil, &result)
		return f.shouldRetry(ctx, resp, err)
	})
	if err != nil {
		return "", err
	}

	if result.ErrorCode != 0 {
		return "", &result.Error
	}

	return result.Link, nil
}

// Check the interfaces are satisfied
var (
	_ fs.Fs           = (*Fs)(nil)
	_ fs.Mover        = (*Fs)(nil)
	_ fs.Copier       = (*Fs)(nil)
	_ fs.Abouter      = (*Fs)(nil)
	_ fs.PublicLinker = (*Fs)(nil)
	_ fs.Object       = (*Object)(nil)
	_ fs.IDer         = (*Object)(nil)
)
