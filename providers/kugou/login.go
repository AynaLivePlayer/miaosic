package kugou

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/tidwall/gjson"
	"time"
)

func (n *Kugou) Login(username string, password string) error {
	return miaosic.ErrNotImplemented
}

func (n *Kugou) IsLogin() bool {
	return false
}

func (n *Kugou) QrLogin() (*miaosic.QrLoginSession, error) {
	unixMili := time.Now().UnixMilli()
	data := map[string]string{
		"appid":      n.appid,
		"clientver":  n.clientver,
		"clienttime": fmt.Sprintf("%d", unixMili),
		"mid":        getMD5Hash(n.dfid),
		"uuid":       getMD5Hash(n.dfid),
		"dfid":       n.dfid,
		"type":       "1",
		"plat":       "4",
		"qrcode_txt": "https://h5.kugou.com/apps/loginQRCode/html/index.html?appid=" + n.appid + "&",
		"srcappid":   "2919",
	}
	data["signature"] = signatureWebParams(data)
	resp, err := miaosic.Requester.GetQueryX("http://login-user.kugou.com/v2/qrcode", data)
	if err != nil {
		return nil, err
	}
	respJson := gjson.ParseBytes(resp.Body())
	key := respJson.Get("data.qrcode").String()
	if key == "" {
		return nil, errors.New("kugou: failed to get qrlogin key")
	}
	return &miaosic.QrLoginSession{
		Url: "https://h5.kugou.com/apps/loginQRCode/html/index.html?qrcode=" + key,
		Key: key,
	}, nil
}

func (n *Kugou) QrLoginVerify(qrlogin *miaosic.QrLoginSession) (*miaosic.QrLoginResult, error) {
	unixMili := time.Now().UnixMilli()
	data := map[string]string{
		"appid":      n.appid,
		"clientver":  n.clientver,
		"clienttime": fmt.Sprintf("%d", unixMili),
		"mid":        getMD5Hash(n.dfid),
		"uuid":       getMD5Hash(n.dfid),
		"dfid":       n.dfid,
		"plat":       "4",
		"qrcode":     qrlogin.Key,
		"srcappid":   "2919",
	}
	data["signature"] = signatureWebParams(data)
	resp, err := miaosic.Requester.GetQueryX("http://login-user.kugou.com/v2/get_userinfo_qrcode", data)
	if err != nil {
		return &miaosic.QrLoginResult{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	respJson := gjson.ParseBytes(resp.Body())
	if respJson.Get("data.status").Int() != 4 {
		return &miaosic.QrLoginResult{
			Success: false,
			Message: "haven't scanned",
		}, nil
	}
	n.cookie["token"] = respJson.Get("data.token").String()
	n.cookie["userid"] = respJson.Get("data.userid").String()
	return &miaosic.QrLoginResult{
		Success: true,
		Message: "login success",
	}, nil
}

func (n *Kugou) Logout() error {
	return nil
}

func (n *Kugou) SaveSession() string {
	b, _ := json.Marshal(n.cookie)
	return base64.StdEncoding.EncodeToString(b)
}

func (n *Kugou) RestoreSession(session string) error {
	b, err := base64.StdEncoding.DecodeString(session)
	if err != nil {
		return err
	}
	data := make(map[string]string)
	err = json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	if token, ok := data["token"]; ok {
		n.cookie["token"] = token
	}
	if userid, ok := data["userid"]; ok {
		n.cookie["userid"] = userid
	}
	return nil
}
