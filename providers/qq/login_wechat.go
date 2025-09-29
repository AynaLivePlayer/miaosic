package qq

import (
	"errors"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cast"
	_ "image/jpeg" // wechat qrcode is jpg
	"net/http"
	"regexp"
	"time"
)

func (p *QQMusicProvider) getWxQR() (*miaosic.QrLoginSession, error) {
	resp, err := miaosic.Requester.GetQuery(
		"https://open.weixin.qq.com/connect/qrconnect",
		map[string]string{
			"appid":         "wx48db31d50e334801",
			"redirect_uri":  "https://y.qq.com/portal/wx_redirect.html?login_type=2&surl=https://y.qq.com/",
			"response_type": "code",
			"scope":         "snsapi_login",
			"state":         "STATE",
			"href":          "https://y.qq.com/mediastyle/music_v17/src/css/popup_wechat.css#wechat_redirect",
		},
		map[string]string{
			"Referer": "https://open.weixin.qq.com/connect/qrconnect",
		},
	)
	if err != nil {
		return nil, err
	}

	loginUuid := regexp.MustCompile("uuid=(.+?)\"").FindStringSubmatch(resp.String())
	if len(loginUuid) < 2 {
		return nil, errors.New("miaosic (qq): failed to get qrcode")
	}
	resp, err = miaosic.Requester.GetQuery(
		"https://open.weixin.qq.com/connect/qrcode/"+loginUuid[1],
		nil,
		map[string]string{
			"Referer": "https://open.weixin.qq.com/connect/qrconnect",
		},
	)
	var qrUrl string
	// !!! dont remove, might use in future as a fallback option.
	//{
	//	img, _, err := image.Decode(bytes.NewBuffer(resp.Body()))
	//	if err != nil {
	//		return nil, errors.New("miaosic (qq): failed to read qrcode")
	//	}
	//
	//	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	//	if err != nil {
	//		return nil, errors.New("miaosic (qq): failed to read qrcode to bmp")
	//	}
	//
	//	qrReader := qrcode.NewQRCodeReader()
	//
	//	result, err := qrReader.Decode(bmp, nil)
	//	if err != nil {
	//		return nil, errors.New("miaosic (qq): failed to decode qrcode")
	//	}
	//	qrUrl = result.GetText()
	//}
	{
		qrUrl = "https://open.weixin.qq.com/connect/confirm?uuid=" + loginUuid[1]
	}
	return &miaosic.QrLoginSession{
		Url: qrUrl,
		Key: loginUuid[1],
	}, nil
}

func (p *QQMusicProvider) checkWxQR(qrlogin *miaosic.QrLoginSession) (*miaosic.QrLoginResult, error) {
	resp, err := resty.New().SetTimeout(time.Second * 2).
		R().
		SetQueryParams(map[string]string{
			"uuid": qrlogin.Key,
			"_":    fmt.Sprintf("%d", time.Now().UnixMilli()),
		}).
		SetHeaders(map[string]string{
			"Referer": "https://open.weixin.qq.com/",
		}).Get("https://lp.open.weixin.qq.com/connect/l/qrconnect")
	if err != nil {
		if errors.Is(err, http.ErrHandlerTimeout) {
			return &miaosic.QrLoginResult{Success: false, Message: "timeout, might be waiting for scan"}, nil
		}
		return &miaosic.QrLoginResult{Success: false, Message: "unknown error"}, err
	}

	//pp.Println(resp.String())

	result := regexp.MustCompile(`window\.wx_errcode=(\d+);window\.wx_code=\'([^\']*)\'`).FindStringSubmatch(resp.String())
	if len(result) < 3 {
		return &miaosic.QrLoginResult{
			Success: false,
			Message: "fail to check qr status",
		}, errors.New("miaosic (qq): fail to check qr status")
	}
	//pp.Println(result)

	statusCode, err := cast.ToIntE(result[1])
	if err != nil {
		return &miaosic.QrLoginResult{
			Success: false,
			Message: "invalid status code",
		}, fmt.Errorf("miaosic (qq): invalid qr status code: %s", result[1])
	}
	switch statusCode {
	case 405:
		return p.getCredentialWithCode(result[2], 1)
	case 408:
		return &miaosic.QrLoginResult{Success: false, Message: "等待扫描二维码"}, nil
	case 404:
		return &miaosic.QrLoginResult{Success: false, Message: "扫描未确认登陆"}, nil
	//case 65:
	//	return &miaosic.QrLoginResult{Success: false, Message: "二维码已过期"}, nil
	case 403:
		return &miaosic.QrLoginResult{Success: false, Message: "! 拒绝登陆 !"}, nil
	default:
		return &miaosic.QrLoginResult{Success: false, Message: fmt.Sprintf("未知错误 %d", statusCode)}, nil
	}
}
