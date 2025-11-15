// Package api provides types for the Baidu Pan API
package api

import (
	"fmt"
	"time"
)

const (
	// API endpoints
	APIBaseURL   = "https://pan.baidu.com"
	OpenAPIURL   = "https://openapi.baidu.com"
	AuthURL      = "https://openapi.baidu.com/oauth/2.0/authorize"
	TokenURL     = "https://openapi.baidu.com/oauth/2.0/token"
	DeviceURL    = "https://openapi.baidu.com/oauth/2.0/device/code"
	
	// API paths
	PathFileList     = "/rest/2.0/xpan/file"
	PathMultimedia   = "/rest/2.0/xpan/multimedia"
	PathFileManager  = "/rest/2.0/xpan/file"
	PathQuota        = "/api/quota"
	PathNAS          = "/rest/2.0/xpan/nas"
	PathPrecreate    = "/rest/2.0/xpan/file"
	PathUpload       = "/rest/2.0/pcs/superfile2"
	PathCreate       = "/rest/2.0/xpan/file"
	
	// File types
	FileTypeFile   = 0
	FileTypeFolder = 1
)

// Error represents an API error response
type Error struct {
	ErrorCode int         `json:"errno"`
	ErrorMsg  string      `json:"errmsg"`
	RequestID interface{} `json:"request_id"` // Can be string or int64
}

// Error implements the error interface
func (e *Error) Error() string {
	return fmt.Sprintf("baidu pan error %d: %s (request_id: %v)", e.ErrorCode, e.ErrorMsg, e.RequestID)
}

// UserInfo represents user information
type UserInfo struct {
	BaiduName  string `json:"baidu_name"`
	NetdiskName string `json:"netdisk_name"`
	AvatarURL   string `json:"avatar_url"`
	VIPType     int    `json:"vip_type"`
	UK          int64  `json:"uk"`
}

// QuotaInfo represents quota information
type QuotaInfo struct {
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
	Free  int64 `json:"free"`
}

// FileListResponse represents the response from file list API
type FileListResponse struct {
	Error
	List   []FileItem `json:"list"`
	HasMore int       `json:"has_more"`
	Cursor int64     `json:"cursor"`
}

// FileItem represents a file or directory
type FileItem struct {
	FsID         int64  `json:"fs_id"`
	Path         string `json:"path"`
	ServerFilename string `json:"server_filename"`
	Size         int64  `json:"size"`
	ServerMtime  int64  `json:"server_mtime"`  // modification time
	ServerCtime  int64  `json:"server_ctime"`  // creation time
	LocalMtime   int64  `json:"local_mtime"`
	LocalCtime   int64  `json:"local_ctime"`
	Isdir        int    `json:"isdir"`  // 0: file, 1: directory
	Category     int    `json:"category"`
	MD5          string `json:"md5"`
	DirEmpty     int    `json:"dir_empty"`
	Thumbs       map[string]string `json:"thumbs"`
}

// IsDir returns true if the item is a directory
func (f *FileItem) IsDir() bool {
	return f.Isdir == FileTypeFolder
}

// ModTime returns the modification time
func (f *FileItem) ModTime() time.Time {
	return time.Unix(f.ServerMtime, 0)
}

// FileMetasRequest represents the request for file metadata
type FileMetasRequest struct {
	Fsids  []int64 `json:"fsids"`
	Dlink  int     `json:"dlink"`  // 1: get download link
}

// FileMetasResponse represents the response from file metas API
type FileMetasResponse struct {
	Error
	List []FileMeta `json:"list"`
}

// FileMeta represents file metadata including download link
type FileMeta struct {
	FsID         int64  `json:"fs_id"`
	Path         string `json:"path"`
	ServerFilename string `json:"server_filename"`
	Size         int64  `json:"size"`
	ServerMtime  int64  `json:"server_mtime"`
	ServerCtime  int64  `json:"server_ctime"`
	LocalMtime   int64  `json:"local_mtime"`
	LocalCtime   int64  `json:"local_ctime"`
	Isdir        int    `json:"isdir"`
	Category     int    `json:"category"`
	MD5          string `json:"md5"`
	Dlink        string `json:"dlink"`  // download link
}

// CreateDirResponse represents the response from create directory API
type CreateDirResponse struct {
	Error
	FsID   int64  `json:"fs_id"`
	Path   string `json:"path"`
	Ctime  int64  `json:"ctime"`
	Mtime  int64  `json:"mtime"`
	Name   string `json:"name"`
}

// FileManagerRequest represents the request for file manager operations
type FileManagerRequest struct {
	Opera   string         `json:"opera"`  // operation: move, copy, delete, rename
	FileList []FileOpItem  `json:"filelist"`
	Async   int            `json:"async"`  // 0: sync, 1: async
	Ondup   string         `json:"ondup,omitempty"`  // overwrite, newcopy, fail
}

// FileOpItem represents a file operation item
type FileOpItem struct {
	Path    string `json:"path"`
	Dest    string `json:"dest,omitempty"`
	NewName string `json:"newname,omitempty"`
}

// FileManagerResponse represents the response from file manager API
type FileManagerResponse struct {
	Error
	TaskID int64          `json:"taskid,omitempty"`
	Info   []FileOpInfo   `json:"info,omitempty"`
}

// FileOpInfo represents file operation result info
type FileOpInfo struct {
	Path string `json:"path"`
}

// PrecreateRequest represents the request for precreate API
type PrecreateRequest struct {
	Path         string   `json:"path"`
	Size         int64    `json:"size"`
	Isdir        int      `json:"isdir"`
	Autoinit     int      `json:"autoinit"`
	BlockList    []string `json:"block_list"`  // MD5 list of each block
	Rtype        int      `json:"rtype"`       // 1: check if file exists, 3: overwrite
	ContentMD5   string   `json:"content-md5,omitempty"`
	SliceMD5     string   `json:"slice-md5,omitempty"`
}

// PrecreateResponse represents the response from precreate API
type PrecreateResponse struct {
	Error
	Path       string `json:"path"`
	UploadID   string `json:"uploadid"`
	ReturnType int    `json:"return_type"`  // 1: need upload, 2: rapid upload success
	BlockList  []int  `json:"block_list"`   // blocks need to be uploaded
}

// UploadResponse represents the response from upload API
type UploadResponse struct {
	Error
	MD5        string `json:"md5"`
	ServerMD5  string `json:"server_md5"`
}

// CreateFileRequest represents the request for create file API
type CreateFileRequest struct {
	Path       string   `json:"path"`
	Size       int64    `json:"size"`
	Isdir      int      `json:"isdir"`
	UploadID   string   `json:"uploadid"`
	BlockList  []string `json:"block_list"`  // MD5 list of uploaded blocks
	Rtype      int      `json:"rtype"`
}

// CreateFileResponse represents the response from create file API
type CreateFileResponse struct {
	Error
	FsID         int64  `json:"fs_id"`
	Path         string `json:"path"`
	ServerFilename string `json:"server_filename"`
	Category     int    `json:"category"`
	Size         int64  `json:"size"`
	MD5          string `json:"md5"`
	ServerMtime  int64  `json:"server_mtime"`
	ServerCtime  int64  `json:"server_ctime"`
	LocalMtime   int64  `json:"local_mtime"`
	LocalCtime   int64  `json:"local_ctime"`
	Isdir        int    `json:"isdir"`
}

// DeviceCodeResponse represents the response from device code API
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	QRCodeURL       string `json:"qrcode_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

