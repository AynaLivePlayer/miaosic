package qq

import "strings"

type Credential struct {
	OpenID       string `json:"openid"`
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	ExpiredAt    int64  `json:"expired_at"`
	MusicID      int64  `json:"musicid"`
	MusicKey     string `json:"musickey"`
	UnionID      string `json:"unionid"`
	StrMusicID   string `json:"str_musicid"`
	RefreshKey   string `json:"refresh_key"`
	EncryptUin   string `json:"encryptUin"`
	LoginType    int    `json:"loginType"`
}

func NewCredential() *Credential {
	return &Credential{}
}

func (c *Credential) GetFormatedLoginType() int {
	if c.LoginType == 0 {
		if c.MusicKey != "" && strings.HasPrefix(c.MusicKey, "W_X") {
			c.LoginType = 1
		} else {
			c.LoginType = 2
		}
	}
	return c.LoginType
}

func (c *Credential) HasMusicID() bool {
	return c.MusicID != 0
}

func (c *Credential) HasMusicKey() bool {
	return c.MusicKey != ""
}
