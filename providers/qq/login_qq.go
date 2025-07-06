package qq

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/spf13/cast"
	"image"
	_ "image/png" // qq qrcode is png
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

func (p *QQMusicProvider) getQQQR() (*miaosic.QrLoginSession, error) {
	resp, err := miaosic.Requester.GetQuery(
		"https://ssl.ptlogin2.qq.com/ptqrshow",
		map[string]string{
			"appid":      "716027609",
			"e":          "2",
			"l":          "M",
			"s":          "3",
			"d":          "72",
			"v":          "4",
			"t":          cast.ToString(rng.Float64()),
			"daid":       "383",
			"pt_3rd_aid": "100497308",
		},
		map[string]string{
			"Referer": "https://ssl.ptlogin2.qq.com/",
		},
	)
	if err != nil {
		return nil, err
	}
	// 获取qrsig cookie
	var qrsig string
	for _, cookie := range resp.RawResponse.Cookies() {
		if cookie.Name == "qrsig" {
			qrsig = cookie.Value
			break
		}
	}

	if qrsig == "" {
		return nil, errors.New("miaosic (qq): failed to get qrsig")
	}

	img, _, err := image.Decode(bytes.NewBuffer(resp.Body()))
	if err != nil {
		return nil, errors.New("miaosic (qq): failed to read qrcode")
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, errors.New("miaosic (qq): failed to read qrcode to bmp")
	}

	qrReader := qrcode.NewQRCodeReader()

	result, err := qrReader.Decode(bmp, nil)
	if err != nil {
		return nil, errors.New("miaosic (qq): failed to decode qrcode")
	}

	return &miaosic.QrLoginSession{
		Url: result.GetText(),
		Key: qrsig,
	}, nil
}

func (p *QQMusicProvider) checkQQQR(qrlogin *miaosic.QrLoginSession) (*miaosic.QrLoginResult, error) {
	ptqrtoken := strconv.Itoa(hash33(qrlogin.Key, 0))
	resp, err := miaosic.Requester.GetQuery(
		"https://ssl.ptlogin2.qq.com/ptqrlogin",
		map[string]string{
			"u1":         "https://graph.qq.com/oauth2.0/login_jump",
			"ptqrtoken":  ptqrtoken,
			"ptredirect": "0",
			"h":          "1",
			"t":          "1",
			"g":          "1",
			"from_ui":    "1",
			"ptlang":     "2052",
			"action":     fmt.Sprintf("0-0-%d", time.Now().UnixMilli()),
			"js_ver":     "20102616",
			"js_type":    "1",
			"pt_uistyle": "40",
			"aid":        "716027609",
			"daid":       "383",
			"pt_3rd_aid": "100497308",
			"has_onekey": "1",
		},
		map[string]string{
			"Referer": "https://xui.ptlogin2.qq.com/",
			"Cookie":  "qrsig=" + qrlogin.Key,
		},
	)
	if err != nil {
		return &miaosic.QrLoginResult{Success: false, Message: "http error, might be invalid qrsig"}, err
	}

	// 使用正则表达式提取状态码
	re := regexp.MustCompile(`ptuiCB\('(\d+)','0','(.*?)','0','(.*?)', (.*?)'\)`)
	matches := re.FindStringSubmatch(resp.String())
	if len(matches) < 5 {
		return &miaosic.QrLoginResult{Success: false, Message: "invalid response"}, errors.New("miaosic (qq): invalid response format")
	}

	statusCode, _ := strconv.Atoi(matches[1])

	//DONE = (0, 405)
	//SCAN = (66, 408)
	//CONF = (67, 404)
	//TIMEOUT = (65, None)
	//REFUSE = (68, 403)
	//OTHER = (None, None)
	switch statusCode {
	case 0:
		return p.authorizeQQQR(qrlogin, matches[2])
	case 66:
		return &miaosic.QrLoginResult{Success: false, Message: "等待扫描二维码"}, nil
	case 67:
		return &miaosic.QrLoginResult{Success: false, Message: "扫描未确认登陆"}, nil
	case 65:
		return &miaosic.QrLoginResult{Success: false, Message: "二维码已过期"}, nil
	case 68:
		return &miaosic.QrLoginResult{Success: false, Message: "! 拒绝登陆 !"}, nil
	default:
		return &miaosic.QrLoginResult{Success: false, Message: matches[3]}, nil
	}
}

func (p *QQMusicProvider) authorizeQQQR(qrlogin *miaosic.QrLoginSession, urlStr string) (*miaosic.QrLoginResult, error) {
	// 解析URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return &miaosic.QrLoginResult{Success: false, Message: "invalid response"}, nil
	}

	// 提取参数
	params := u.Query()
	uin := params.Get("uin")
	sigx := params.Get("ptsigx")
	if uin == "" || sigx == "" {
		return &miaosic.QrLoginResult{Success: false, Message: "invalid response"}, errors.New("miaosic (qq): missing uin or sigx")
	}
	//// 使用 Requester 发送 GET 请求
	//resp, err := miaosic.Requester.GetQuery(
	//	urlStr,
	//	map[string]string{},
	//	map[string]string{
	//		"Referer": "https://xui.ptlogin2.qq.com/",
	//		"Cookie":  "qrsig=" + qrlogin.Key,
	//	},
	//)
	//if err != nil {
	//	return nil, err
	//}

	respR, err := resty.New().
		SetRedirectPolicy(resty.NoRedirectPolicy()).
		R().
		SetHeader("Referer", "https://xui.ptlogin2.qq.com/").
		Get(urlStr)

	if respR == nil {
		return &miaosic.QrLoginResult{Success: false, Message: "invalid response"}, nil
	}

	// 获取p_skey cookie
	var pSkey string
	for _, cookie := range respR.RawResponse.Cookies() {
		if cookie.Name == "p_skey" {
			pSkey = cookie.Value
			break
		}
	}

	if pSkey == "" {
		return &miaosic.QrLoginResult{Success: false, Message: "failed to get p_skey"}, errors.New("miaosic (qq): failed to get p_skey")
	}

	gTk := hash33(pSkey, 5381)

	// 构建表单数据
	formData := url.Values{}
	formData.Set("response_type", "code")
	formData.Set("client_id", "100497308")
	formData.Set("redirect_uri", "https://y.qq.com/portal/wx_redirect.html?login_type=1&surl=https://y.qq.com/")
	formData.Set("scope", "get_user_info,get_app_friends")
	formData.Set("state", "state")
	formData.Set("switch", "")
	formData.Set("from_ptlogin", "1")
	formData.Set("src", "1")
	formData.Set("update_auth", "1")
	formData.Set("openapi", "1010_1030")
	formData.Set("g_tk", strconv.Itoa(gTk))
	formData.Set("auth_time", strconv.FormatInt(time.Now().UnixMilli(), 10))
	formData.Set("ui", uuid.New().String())

	respR, err = resty.New().
		SetRedirectPolicy(resty.NoRedirectPolicy()).
		R().
		SetCookies(respR.Cookies()).
		SetFormDataFromValues(formData).
		SetHeaders(map[string]string{
			"Referer": "https://xui.ptlogin2.qq.com/",
		}).Post("https://graph.qq.com/oauth2.0/authorize")

	//resp, err := miaosic.Requester.Post(
	//	"https://graph.qq.com/oauth2.0/authorize",
	//	map[string]string{
	//		"Content-Type": "application/x-www-form-urlencoded",
	//		"Referer":      "https://xui.ptlogin2.qq.com/",
	//	},
	//	formData.Encode(),
	//)
	if respR == nil {
		return &miaosic.QrLoginResult{Success: false, Message: "oauth failed, access failed"}, nil
	}

	location := respR.Header().Get("Location")
	if location == "" {
		return &miaosic.QrLoginResult{Success: false, Message: "oauth failed, no location found"}, nil
	}

	u, err = url.Parse(location)
	if err != nil {
		return nil, err
	}

	code := u.Query().Get("code")
	if code == "" {
		return &miaosic.QrLoginResult{Success: false, Message: "oauth failed, no code in redirection location"}, nil
	}
	return p.getCredentialWithCode(code, 2) // 2 表示QQ登录
}

func (p *QQMusicProvider) getCredentialWithCode(code string, loginType int) (*miaosic.QrLoginResult, error) {
	p.cred.LoginType = loginType
	params := map[string]interface{}{
		"code": code,
	}

	data, err := p.makeApiRequest("QQConnectLogin.LoginServer", "QQLogin", params)
	if err != nil {
		return nil, err
	}

	if !data.Get("data.musickey").Exists() || data.Get("data.musickey").String() == "" {
		return &miaosic.QrLoginResult{Success: false, Message: "fail to get status data"}, nil
	}

	p.cred.OpenID = data.Get("data.openid").String()
	p.cred.RefreshToken = data.Get("data.refresh_token").String()
	p.cred.AccessToken = data.Get("data.access_token").String()
	p.cred.ExpiredAt = data.Get("data.expired_at").Int()
	p.cred.MusicID = data.Get("data.musicid").Int()
	p.cred.MusicKey = data.Get("data.musickey").String()
	p.cred.UnionID = data.Get("data.unionid").String()
	p.cred.StrMusicID = data.Get("data.str_musicid").String()
	p.cred.RefreshKey = data.Get("data.refresh_key").String()
	p.cred.EncryptUin = data.Get("data.encryptUin").String()
	p.cred.LoginType = int(data.Get("data.loginType").Int())
	return &miaosic.QrLoginResult{Success: true, Message: "ok"}, nil
}
