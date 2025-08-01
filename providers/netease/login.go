package netease

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	neteaseApi "github.com/XiaoMengXinX/Music163Api-Go/api"
	"net/http"
)

func (n *Netease) Login(username string, password string) error {
	return miaosic.ErrNotImplemented
}

func (n *Netease) IsLogin() bool {
	status, err := neteaseApi.GetLoginStatus(n.ReqData)
	if err != nil {
		return false
	}
	return status.Account.Id != 0
}

func (p *Netease) RefreshLogin() error {
	return nil
}

func (n *Netease) QrLogin() (*miaosic.QrLoginSession, error) {
	unikey, err := neteaseApi.GetQrUnikey(n.ReqData)
	if err != nil {
		return nil, err
	}
	return &miaosic.QrLoginSession{
		Key: unikey.Unikey,
		Url: fmt.Sprintf("https://music.163.com/login?codekey=%s", unikey.Unikey),
	}, nil
}

func (n *Netease) QrLoginVerify(qrlogin *miaosic.QrLoginSession) (*miaosic.QrLoginResult, error) {
	login, h, err := neteaseApi.CheckQrLogin(n.ReqData, qrlogin.Key)
	if err != nil {
		return nil, err
	}
	// if login.Code == 800 || login.Code == 803. login success
	if login.Code != 800 && login.Code != 803 {
		return &miaosic.QrLoginResult{
			Success: false,
			Message: login.Message,
		}, nil
	}
	cookies := make([]*http.Cookie, 0)
	for _, c := range (&http.Response{Header: h}).Cookies() {
		if c.Name == "MUSIC_U" || c.Name == "__csrf" || c.Name == "MUSIC_A" {
			cookies = append(cookies, c)
		}
	}
	cookies = append(cookies, &http.Cookie{
		Name:  "deviceId",
		Value: n.deviceId,
	})
	n.ReqData.Cookies = cookies
	return &miaosic.QrLoginResult{
		Success: true,
		Message: login.Message,
	}, nil
}

func (n *Netease) Logout() error {
	n.ReqData.Cookies = []*http.Cookie{
		{Name: "MUSIC_U", Value: ""},
		{Name: "__csrf", Value: ""},
		{Name: "deviceId", Value: n.deviceId},
	}
	return nil
}

func (n *Netease) SaveSession() string {
	// save session to string MUSIC_U and __csrf
	data := make(map[string]string)
	data["MUSIC_U"] = ""
	data["__csrf"] = ""
	for _, c := range n.ReqData.Cookies {
		if c.Name == "MUSIC_U" || c.Name == "MUSIC_A" || c.Name == "__csrf" || c.Name == "deviceId" {
			data[c.Name] = c.Value
		}
	}
	data["deviceId"] = n.deviceId
	b, _ := json.Marshal(data)
	return base64.StdEncoding.EncodeToString(b)
}

func (n *Netease) RestoreSession(session string) error {
	// restore session from string MUSIC_U and __csrf
	b, err := base64.StdEncoding.DecodeString(session)
	if err != nil {
		return err
	}
	data := make(map[string]string)
	err = json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	cookies := make([]*http.Cookie, 0)
	for name, value := range data {
		if name == "MUSIC_U" || name == "MUSIC_A" || name == "__csrf" {
			cookies = append(cookies, &http.Cookie{
				Name:  name,
				Value: value,
			})
		}
		if name == "deviceId" {
			n.deviceId = value
		}
	}
	cookies = append(cookies, &http.Cookie{
		Name:  "deviceId",
		Value: n.deviceId,
	})
	n.ReqData.Cookies = cookies
	return nil
}
