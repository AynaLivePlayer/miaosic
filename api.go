package miaosic

func SearchByProvider(provider string, keyword string, page, size int) ([]MediaInfo, error) {
	p, ok := GetProvider(provider)
	if !ok {
		return nil, ErrorNoSuchProvider
	}
	return p.Search(keyword, page, size)
}

func GetMediaUrl(meta MetaData, quality Quality) ([]MediaUrl, error) {
	provider, ok := GetProvider(meta.Provider)
	if !ok {
		return nil, ErrorNoSuchProvider
	}
	return provider.GetMediaUrl(meta, quality)
}

func GetMediaInfo(meta MetaData) (MediaInfo, error) {
	provider, ok := GetProvider(meta.Provider)
	if !ok {
		return MediaInfo{}, ErrorNoSuchProvider
	}
	return provider.GetMediaInfo(meta)
}

func GetMediaLyric(meta MetaData) ([]Lyrics, error) {
	provider, ok := GetProvider(meta.Provider)
	if !ok {
		return nil, ErrorNoSuchProvider
	}
	return provider.GetMediaLyric(meta)
}

func MatchPlaylistByProvider(provider string, uri string) (MetaData, bool) {
	p, ok := GetProvider(provider)
	if !ok {
		return MetaData{}, false
	}
	return p.MatchPlaylist(uri)
}

func GetPlaylist(meta MetaData) (*Playlist, error) {
	p, ok := GetProvider(meta.Provider)
	if !ok {
		return nil, ErrorNoSuchProvider
	}
	return p.GetPlaylist(meta)
}

func MatchMedia(keyword string) (MetaData, bool) {
	for _, p := range _providers {
		if meta, ok := p.MatchMedia(keyword); ok {
			return meta, true
		}
	}
	return MetaData{}, false
}

func MatchMediaByProvider(provider string, uri string) (MetaData, bool) {
	p, ok := GetProvider(provider)
	if !ok {
		return MetaData{}, false
	}
	return p.MatchMedia(uri)
}

func loginableByProvider(provider string) (Loginable, error) {
	p, ok := GetProvider(provider)
	if !ok {
		return nil, ErrorNoSuchProvider
	}
	loginable, ok := p.(Loginable)
	if !ok {
		return nil, ErrorProviderNotLoginable
	}
	return loginable, nil
}

func LoginByProvider(provider, username, password string) error {
	loginable, err := loginableByProvider(provider)
	if err != nil {
		return err
	}
	return loginable.Login(username, password)
}

func LogoutByProvider(provider string) error {
	loginable, err := loginableByProvider(provider)
	if err != nil {
		return err
	}
	return loginable.Logout()
}

func IsLoginByProvider(provider string) (bool, error) {
	loginable, err := loginableByProvider(provider)
	if err != nil {
		return false, err
	}
	return loginable.IsLogin(), nil
}

func RefreshLoginByProvider(provider string) error {
	loginable, err := loginableByProvider(provider)
	if err != nil {
		return err
	}
	return loginable.RefreshLogin()
}

func QrLoginByProvider(provider string) (*QrLoginSession, error) {
	loginable, err := loginableByProvider(provider)
	if err != nil {
		return nil, err
	}
	return loginable.QrLogin()
}

func QrLoginVerifyByProvider(provider string, qrlogin *QrLoginSession) (*QrLoginResult, error) {
	loginable, err := loginableByProvider(provider)
	if err != nil {
		return nil, err
	}
	return loginable.QrLoginVerify(qrlogin)
}

func RestoreSessionByProvider(provider, session string) error {
	loginable, err := loginableByProvider(provider)
	if err != nil {
		return err
	}
	return loginable.RestoreSession(session)
}

func SaveSessionByProvider(provider string) (string, error) {
	loginable, err := loginableByProvider(provider)
	if err != nil {
		return "", err
	}
	return loginable.SaveSession(), nil
}
