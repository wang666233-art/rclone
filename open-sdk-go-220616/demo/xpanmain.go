package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	openapiclient "open-sdk-go/openxpanapi"
)

const (
	MyAccessToken = "123.565dff38db8ca0593f752303f49ba413.Ysjdy668QCK3B02hotvPZRQoIpog6a1dHGOTMmO.QIOJZQ"
)

type precreateReturnType struct {
	Path       string        `json:"path"`
	Uploadid   string        `json:"uploadid"`
	ReturnType int           `json:"return_type"`
	BlockList  []interface{} `json:"block_list"`
	Errno      int           `json:"errno"`
	RequestID  int64         `json:"request_id"`
}

type authReturnType struct {
	ExpiresIn     int    `json:"expires_in"`
	RefreshToken  string `json:"refresh_token"`
	AccessToken   string `json:"access_token"`
	SessionSecret string `json:"session_secret"`
	SessionKey    string `json:"session_key"`
	Scope         string `json:"scope"`
}

type deviceReturnType struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	QrcodeURL       string `json:"qrcode_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func MyXpanfileprecreate() string {
	accessToken := MyAccessToken                          // string
	path := "/来自：小度设备/a.txt"                              // string
	isdir := int32(0)                                     // int32
	size := int32(14)                                     // int32
	autoinit := int32(1)                                  // int32
	blockList := "[\"f649cdca4f6987c74b0471e87f195af4\"]" // string
	rtype := int32(3)                                     // int32 | rtype (optional)

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.FileuploadApi.Xpanfileprecreate(context.Background()).AccessToken(accessToken).Path(path).Isdir(isdir).Size(size).Autoinit(autoinit).BlockList(blockList).Rtype(rtype).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileuploadApi.Xpanfileprecreate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Xpanfileprecreate`: Fileprecreateresponse
	fmt.Fprintf(os.Stdout, "Response from `FileuploadApi.Xpanfileprecreate`: %v\n", resp)

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}

	fmt.Println(string(bodyBytes))
	return string(bodyBytes)
}

func MyPcssuperfile2(uplodidTmp string) {
	accessToken := MyAccessToken // string
	partseq := "0"               // string
	path := "/来自：小度设备/a.txt"
	uploadid := uplodidTmp // string
	type_ := "tmpfile"
	file, err := os.Open("./demo/data/1.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	configuration := openapiclient.NewConfiguration()
	//configuration.Debug = true
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.FileuploadApi.Pcssuperfile2(context.Background()).AccessToken(accessToken).Partseq(partseq).Path(path).Uploadid(uploadid).Type_(type_).File(file).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileuploadApi.Pcssuperfile2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Pcssuperfile2`: string
	fmt.Fprintf(os.Stdout, "Response from `FileuploadApi.Pcssuperfile2`: %v\n", resp)

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}

	fmt.Println(string(bodyBytes))
}

func MyXpanfilecreate(uplodidTmp string) {
	accessToken := MyAccessToken // string
	path := "/来自：小度设备/a.txt"
	isdir := int32(0)                                     // int32
	size := int32(14)                                     // int32
	uploadid := uplodidTmp                                // string
	blockList := "[\"f649cdca4f6987c74b0471e87f195af4\"]" // string
	rtype := int32(3)                                     // int32 | rtype (optional)

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.FileuploadApi.Xpanfilecreate(context.Background()).AccessToken(accessToken).Path(path).Isdir(isdir).Size(size).Uploadid(uploadid).BlockList(blockList).Rtype(rtype).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileuploadApi.Xpanfilecreate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Xpanfilecreate`: Filecreateresponse
	fmt.Fprintf(os.Stdout, "Response from `FileuploadApi.Xpanfilecreate`: %v\n", resp)

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}

	fmt.Println(string(bodyBytes))
}

func MyXpanfileimagelist() {
	accessToken := MyAccessToken // string
	web := "1"                   // string |  (optional)
	parentPath := "/"            // string |  (optional)
	recursion := "1"             // string |  (optional)
	page := int32(1)             // int32 |  (optional)
	num := int32(1)              // int32 |  (optional)
	order := "time"              // string |  (optional)
	desc := "1"                  // string |  (optional)

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.FileinfoApi.Xpanfileimagelist(context.Background()).AccessToken(accessToken).ParentPath(parentPath).Recursion(recursion).Page(page).Num(num).Order(order).Desc(desc).Web(web).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileinfoApi.Xpanfileimagelist``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}
	bodyString := string(bodyBytes)

	fmt.Fprintf(os.Stdout, "Response from `FileinfoApi.Xpanfileimagelist`: %v body: %v\n", r, bodyString)
	fmt.Fprintf(os.Stdout, "Response from `FileinfoApi.Xpanfileimagelist`: %v\n", resp)

}

func MyXpanfiledoclist() {
	accessToken := MyAccessToken // string
	web := "1"                   // string |  (optional)
	parentPath := "/"            // string |  (optional)
	page := int32(1)             // int32 |  (optional)
	num := int32(1)              // int32 |  (optional)
	recursion := "1"             // string |  (optional)
	order := "time"              // string |  (optional)
	desc := "1"                  // string |  (optional)

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.FileinfoApi.Xpanfiledoclist(context.Background()).AccessToken(accessToken).ParentPath(parentPath).Recursion(recursion).Page(page).Num(num).Order(order).Desc(desc).Web(web).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileinfoApi.Xpanfiledoclist``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}
	bodyString := string(bodyBytes)

	fmt.Fprintf(os.Stdout, "Response from `FileinfoApi.Xpanfiledoclist`: %v body: %v\n", r, bodyString)
	fmt.Fprintf(os.Stdout, "Response from `FileinfoApi.Xpanfiledoclist`: %v\n", resp)
}

func MyXpanfilesearch() {
	accessToken := MyAccessToken // string |
	web := "1"                   // string |  (optional)
	num := "2"                   // string |  (optional)
	page := "1"                  // string |  (optional)
	dir := "/"                   // string |  (optional)
	recursion := "1"             // string |  (optional)
	key := "txt"                 // string |

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.FileinfoApi.Xpanfilesearch(context.Background()).AccessToken(accessToken).Web(web).Num(num).Page(page).Dir(dir).Recursion(recursion).Key(key).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileinfoApi.Xpanfilesearch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}
	bodyString := string(bodyBytes)

	fmt.Fprintf(os.Stdout, "Response from `FileinfoApi.Xpanfilesearch`: %v body: %v\n", r, bodyString)
	fmt.Fprintf(os.Stdout, "Response from `FileinfoApi.Xpanfilesearch`: %v\n", resp)
}

func MyXpanfilelist() {
	accessToken := MyAccessToken // string |
	folder := "0"                // string |  (optional)
	web := "web"                 // string |  (optional)
	start := "0"                 // string |  (optional)
	limit := int32(2)            // int32 |  (optional)
	dir := "/"                   // string |  (optional)
	order := "time"              // string |  (optional)
	desc := int32(1)             // int32 |  (optional)

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.FileinfoApi.Xpanfilelist(context.Background()).AccessToken(accessToken).Folder(folder).Web(web).Start(start).Limit(limit).Dir(dir).Order(order).Desc(desc).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileinfoApi.Xpanfilelist``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	fmt.Fprintf(os.Stdout, "Response from `FileinfoApi.Xpanfilelist`: %v\n", resp)
}

func MyApiQuota() {
	accessToken := MyAccessToken // string
	checkexpire := int32(1)      // int32 |  (optional)
	checkfree := int32(1)        // int32 |  (optional)

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.UserinfoApi.Apiquota(context.Background()).AccessToken(accessToken).Checkexpire(checkexpire).Checkfree(checkfree).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FileinfoApi.Xpanfilelist``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ApiQuotaGet`: string
	fmt.Fprintf(os.Stdout, "Response from `FileinfoApi.Xpanfilelist`: %v\n", resp)
}

func MyXpannasuinfo() {
	accessToken := MyAccessToken // string

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.UserinfoApi.Xpannasuinfo(context.Background()).AccessToken(accessToken).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserinfoApi.Xpannasuinfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}
	bodyString := string(bodyBytes)

	fmt.Fprintf(os.Stdout, "Response from `UserinfoApi.Xpannasuinfo`: %v body: %v\n", r, bodyString)
	fmt.Fprintf(os.Stdout, "Response from `UserinfoApi.Xpannasuinfo`: %v\n", resp)
}

func MyXpanmultimediafilemetas() {
	accessToken := MyAccessToken // string
	thumb := "1"                 // string |  (optional)
	extra := "1"                 // string |  (optional)
	fsids := "[992899941915887]" // string
	dlink := "1"                 // string |  (optional)

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.MultimediafileApi.Xpanmultimediafilemetas(context.Background()).AccessToken(accessToken).Thumb(thumb).Extra(extra).Fsids(fsids).Dlink(dlink).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MultimediafileApi.Xpanmultimediafilemetas``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}
	bodyString := string(bodyBytes)

	fmt.Fprintf(os.Stdout, "Response from `MultimediafileApi.Xpanmultimediafilemetas`: %v body: %v\n", r, bodyString)
	fmt.Fprintf(os.Stdout, "Response from `MultimediafileApi.Xpanmultimediafilemetas`: %v\n", resp)
}

func MyXpanfilelistall() {
	accessToken := MyAccessToken // string
	path := "/"                  // string
	web := "1"                   // string |  (optional)
	start := int32(0)            // int32 |  (optional)
	limit := int32(5)            // int32 |  (optional)
	recursion := int32(1)        // int32 | (optional)

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.MultimediafileApi.Xpanfilelistall(context.Background()).AccessToken(accessToken).Path(path).Web(web).Start(start).Limit(limit).Recursion(recursion).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MultimediafileApi.Xpanfilelistall``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}
	bodyString := string(bodyBytes)

	fmt.Fprintf(os.Stdout, "Response from `MultimediafileApi.Xpanfilelistall`: %v body: %v\n", r, bodyString)
	fmt.Fprintf(os.Stdout, "Response from `MultimediafileApi.Xpanfilelistall`: %v\n", resp)
}

func MyOauthTokenAuthorizationCode() string {
	code := "6e5f9cea1d65c8173f0133ae05a09bbd"         // string
	clientId := "QyUVtHyEGny6M6pnSvG94w2duhbDm7Ei"     // string
	clientSecret := "HyE4oaCRH2O8aOMf6Kwxe2wx1A3mGv7Q" // string
	redirectUri := "oob"                               // string

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.AuthApi.OauthTokenCode2token(context.Background()).Code(code).ClientId(clientId).ClientSecret(clientSecret).RedirectUri(redirectUri).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthApi.OauthTokenCode2token``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `OauthTokenCode2token`: OauthTokenAuthorizationCodeResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthApi.OauthTokenCode2token`: %v\n", resp)

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}
	return string(bodyBytes)
}

func MyOauthTokenRefreshToken(rToken string) {
	refreshToken := rToken
	clientId := "QyUVtHyEGny6M6pnSvG94w2duhbDm7Ei"     // string
	clientSecret := "HyE4oaCRH2O8aOMf6Kwxe2wx1A3mGv7Q" // string

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.AuthApi.OauthTokenRefreshToken(context.Background()).RefreshToken(refreshToken).ClientId(clientId).ClientSecret(clientSecret).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthApi.OauthTokenRefreshToken``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `OauthTokenRefreshToken`: OauthTokenRefreshTokenResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthApi.OauthTokenRefreshToken`: %v %v\n", resp, r)
}

func MyOauthTokenDeviceCode() {
	clientId := "QyUVtHyEGny6M6pnSvG94w2duhbDm7Ei" // string
	scope := "basic,netdisk"                       // string

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.AuthApi.OauthTokenDeviceCode(context.Background()).ClientId(clientId).Scope(scope).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthApi.OauthTokenDeviceCode``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `OauthTokenDeviceCode`: OauthTokenDeviceCodeResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthApi.OauthTokenDeviceCode`: %v\n", resp)

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}
	bodyString := string(bodyBytes)

	fmt.Fprintf(os.Stdout, "Response from `AuthApi.OauthTokenDeviceCode`: %v body: %v\n", r, bodyString)
}

func MyOauthTokenDeviceToken(codeTmp string) {
	code := codeTmp                                    // string
	clientId := "QyUVtHyEGny6M6pnSvG94w2duhbDm7Ei"     // string
	clientSecret := "HyE4oaCRH2O8aOMf6Kwxe2wx1A3mGv7Q" // string

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	resp, r, err := api_client.AuthApi.OauthTokenDeviceToken(context.Background()).Code(code).ClientId(clientId).ClientSecret(clientSecret).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthApi.OauthTokenDeviceToken``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `OauthTokenDeviceToken`: OauthTokenDeviceTokenResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthApi.OauthTokenDeviceToken`: %v\n", resp)
}

func MyFilemanagercopy() {
	accessToken := "123.243f175dfe1664651d2fe67fface30be.Y7-t1F0DIy5SI4cXZUktt9LyoRMygzzWpaXkI1Q.hTVwnw"        // string
	async := int32(1)                                                                                           // int32
	filelist := `[{"path":"/test/123123.docx","dest":"/test/abc","newname":"123124.docx","ondup":"overwrite"}]` // string
	ondup := "overwrite"                                                                                        // string | ondup (optional)

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	r, err := api_client.FilemanagerApi.Filemanagercopy(context.Background()).AccessToken(accessToken).Async(async).Filelist(filelist).Ondup(ondup).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilemanagerApi.Filemanagercopy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}
	bodyString := string(bodyBytes)

	fmt.Fprintf(os.Stdout, "Response from `FilemanagerApi.Filemanagercopy`: %v body: %v\n", r, bodyString)

}

func MyFilemanagermove() {
	accessToken := "123.243f175dfe1664651d2fe67fface30be.Y7-t1F0DIy5SI4cXZUktt9LyoRMygzzWpaXkI1Q.hTVwnw"         // string
	async := int32(1)                                                                                            // int32
	filelist := `[{"path":"/test/abc/123124.docx","dest":"/test/","newname":"123124.docx","ondup":"overwrite"}]` // string
	ondup := "overwrite"                                                                                         // string | ondup (optional)

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	r, err := api_client.FilemanagerApi.Filemanagermove(context.Background()).AccessToken(accessToken).Async(async).Filelist(filelist).Ondup(ondup).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilemanagerApi.Filemanagermove``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}
	bodyString := string(bodyBytes)

	fmt.Fprintf(os.Stdout, "Response from `FilemanagerApi.Filemanagermove`: %v body: %v\n", r, bodyString)

}

func MyFilemanagerrename() {
	accessToken := "123.243f175dfe1664651d2fe67fface30be.Y7-t1F0DIy5SI4cXZUktt9LyoRMygzzWpaXkI1Q.hTVwnw" // string
	async := int32(1)                                                                                    // int32
	filelist := `[{"path":"/test/123124.docx","newname":"123125.docx"}]`                                 // string
	ondup := "overwrite"                                                                                 // string | ondup (optional)

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	r, err := api_client.FilemanagerApi.Filemanagerrename(context.Background()).AccessToken(accessToken).Async(async).Filelist(filelist).Ondup(ondup).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilemanagerApi.Filemanagerrename``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}
	bodyString := string(bodyBytes)

	fmt.Fprintf(os.Stdout, "Response from `FilemanagerApi.Filemanagerrename`: %v body: %v\n", r, bodyString)

}

func MyFilemanagerdelete() {
	accessToken := "123.243f175dfe1664651d2fe67fface30be.Y7-t1F0DIy5SI4cXZUktt9LyoRMygzzWpaXkI1Q.hTVwnw" // string
	async := int32(1)                                                                                    // int32
	filelist := `[{"path":"/test/123125.docx"}]`                                                         // string
	ondup := "overwrite"                                                                                 // string | ondup (optional)

	configuration := openapiclient.NewConfiguration()
	api_client := openapiclient.NewAPIClient(configuration)
	r, err := api_client.FilemanagerApi.Filemanagerdelete(context.Background()).AccessToken(accessToken).Async(async).Filelist(filelist).Ondup(ondup).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `FilemanagerApi.Filemanagerdelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", r)
	}
	bodyString := string(bodyBytes)

	fmt.Fprintf(os.Stdout, "Response from `FilemanagerApi.Filemanagerrename`: %v body: %v\n", r, bodyString)

}

func fileupload() {
	// 上传相关的三个接口

	// precreate
	precreateReturnStr := MyXpanfileprecreate()
	p := &precreateReturnType{}
	if err := json.Unmarshal([]byte(precreateReturnStr), p); err != nil {
		fmt.Println(err)
		return
	}

	// superfile2
	MyPcssuperfile2(p.Uploadid)

	// create
	MyXpanfilecreate(p.Uploadid)
}

func filemeta() {
	MyXpanfileimagelist()

	MyXpanfiledoclist()

	MyXpanfilesearch()

	MyXpanfilelist()
}

func userinfo() {
	MyApiQuota()

	MyXpannasuinfo()
}

func multimediafile() {
	MyXpanmultimediafilemetas()

	MyXpanfilelistall()
}

func auth() {
	authCodeStr := MyOauthTokenAuthorizationCode()
	p := &authReturnType{}
	if err := json.Unmarshal([]byte(authCodeStr), p); err != nil {
		fmt.Println(err)
		return
	}
	MyOauthTokenRefreshToken(p.RefreshToken)

	MyOauthTokenDeviceCode()

	MyOauthTokenDeviceToken("53a9e2a414dcb54c5b02b423bad62515")
}

func filemanager() {
	MyFilemanagercopy()

	MyFilemanagermove()

	MyFilemanagerrename()

	MyFilemanagerdelete()
}

func main() {
	fileupload()

	//filemeta()

	//userinfo()

	//multimediafile()

	//auth()

	//filemanager()
}
