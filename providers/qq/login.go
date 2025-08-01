package qq

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"time"
)

func (p *QQMusicProvider) Login(username string, password string) error {
	return miaosic.ErrNotImplemented
}

func (p *QQMusicProvider) Logout() error {
	// todo api request
	p.cred = NewCredential()
	return nil
}

func (p *QQMusicProvider) IsLogin() bool {
	// todo check if token expires
	return p.cred.HasMusicID() && p.cred.HasMusicKey()
}

func (p *QQMusicProvider) RefreshLogin() error {
	return p.refreshToken()
}

func (p *QQMusicProvider) refreshToken() error {
	if p.cred.RefreshKey == "" || p.cred.RefreshToken == "" || !p.cred.HasMusicKey() || !p.cred.HasMusicKey() {
		return errors.New("miaosic (qq): invalid credentials")
	}
	params := map[string]interface{}{
		"refresh_key":   p.cred.RefreshKey,
		"refresh_token": p.cred.RefreshToken,
		"musickey":      p.cred.MusicKey,
		"musicid":       p.cred.MusicID,
	}
	data, err := p.makeApiRequest("music.login.LoginServer",
		"Login", params)
	if err != nil {
		return err
	}
	if !data.Get("data.musickey").Exists() || data.Get("data.musickey").String() == "" {
		return errors.New("miaosic (qq): fail to get login status data")
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
	p.cred.CreatedAt = time.Now().Unix()
	return nil
}

func (p *QQMusicProvider) QrLogin() (*miaosic.QrLoginSession, error) {
	// todo finish wechat qrlogin channel
	return p.getQQQR()
}

func (p *QQMusicProvider) QrLoginVerify(qrlogin *miaosic.QrLoginSession) (*miaosic.QrLoginResult, error) {
	// todo finish wechat qrlogin channel
	return p.checkQQQR(qrlogin)
}

func (p *QQMusicProvider) RestoreSession(session string) error {
	if session == "" {
		return errors.New("miaosic (qq): session is empty")
	}

	b, err := base64.StdEncoding.DecodeString(session)
	if err != nil {
		return err
	}

	var data struct {
		Device     *Device     `json:"device"`
		Credential *Credential `json:"credential"`
	}

	err = json.Unmarshal(b, &data)
	if err != nil {
		return fmt.Errorf("miaosic (qq): failed to unmarshal session data: %w", err)
	}

	if data.Device == nil {
		return errors.New("miaosic (qq): missing device info in session")
	}
	if data.Credential == nil {
		return errors.New("miaosic (qq): missing credential info in session")
	}

	p.device = data.Device
	p.cred = data.Credential
	p.qimeiUpdated = false
	return nil
}

func (p *QQMusicProvider) SaveSession() string {
	data := map[string]interface{}{
		"device":     p.device,
		"credential": p.cred,
	}
	val, _ := json.Marshal(data)
	return base64.StdEncoding.EncodeToString(val)
}
