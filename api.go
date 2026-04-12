package miaosic

var DefaultRegistry = NewRegistry()

func RegisterProvider(provider MediaProvider) {
	DefaultRegistry.RegisterProvider(provider)
}

func UnregisterProvider(name string) {
	DefaultRegistry.UnregisterProvider(name)
}

func UnregisterAllProvider() {
	DefaultRegistry.UnregisterAllProvider()
}

func GetProvider(name string) (MediaProvider, bool) {
	return DefaultRegistry.GetProvider(name)
}

func ListAvailableProviders() []string {
	return DefaultRegistry.ListAvailableProviders()
}

func SearchByProvider(provider string, keyword string, page, size int) ([]MediaInfo, error) {
	return DefaultRegistry.SearchByProvider(provider, keyword, page, size)
}

func (r *Registry) SearchByProvider(provider string, keyword string, page, size int) ([]MediaInfo, error) {
	p, ok := r.GetProvider(provider)
	if !ok {
		return nil, ErrorNoSuchProvider
	}
	return p.Search(keyword, page, size)
}

func GetMediaUrl(meta MetaData, quality Quality) ([]MediaUrl, error) {
	return DefaultRegistry.GetMediaUrl(meta, quality)
}

func (r *Registry) GetMediaUrl(meta MetaData, quality Quality) ([]MediaUrl, error) {
	provider, ok := r.GetProvider(meta.Provider)
	if !ok {
		return nil, ErrorNoSuchProvider
	}
	return provider.GetMediaUrl(meta, quality)
}

func GetMediaInfo(meta MetaData) (MediaInfo, error) {
	return DefaultRegistry.GetMediaInfo(meta)
}

func (r *Registry) GetMediaInfo(meta MetaData) (MediaInfo, error) {
	provider, ok := r.GetProvider(meta.Provider)
	if !ok {
		return MediaInfo{}, ErrorNoSuchProvider
	}
	return provider.GetMediaInfo(meta)
}

func GetMediaLyric(meta MetaData) ([]Lyrics, error) {
	return DefaultRegistry.GetMediaLyric(meta)
}

func (r *Registry) GetMediaLyric(meta MetaData) ([]Lyrics, error) {
	provider, ok := r.GetProvider(meta.Provider)
	if !ok {
		return nil, ErrorNoSuchProvider
	}
	return provider.GetMediaLyric(meta)
}

func MatchPlaylistByProvider(provider string, uri string) (MetaData, bool) {
	return DefaultRegistry.MatchPlaylistByProvider(provider, uri)
}

func (r *Registry) MatchPlaylistByProvider(provider string, uri string) (MetaData, bool) {
	p, ok := r.GetProvider(provider)
	if !ok {
		return MetaData{}, false
	}
	return p.MatchPlaylist(uri)
}

func GetPlaylist(meta MetaData) (*Playlist, error) {
	return DefaultRegistry.GetPlaylist(meta)
}

func (r *Registry) GetPlaylist(meta MetaData) (*Playlist, error) {
	p, ok := r.GetProvider(meta.Provider)
	if !ok {
		return nil, ErrorNoSuchProvider
	}
	return p.GetPlaylist(meta)
}

func MatchMedia(keyword string) (MetaData, bool) {
	return DefaultRegistry.MatchMedia(keyword)
}

func (r *Registry) MatchMedia(keyword string) (MetaData, bool) {
	for _, p := range r.providers {
		if meta, ok := p.MatchMedia(keyword); ok {
			return meta, true
		}
	}
	return MetaData{}, false
}

func MatchMediaByProvider(provider string, uri string) (MetaData, bool) {
	return DefaultRegistry.MatchMediaByProvider(provider, uri)
}

func (r *Registry) MatchMediaByProvider(provider string, uri string) (MetaData, bool) {
	p, ok := r.GetProvider(provider)
	if !ok {
		return MetaData{}, false
	}
	return p.MatchMedia(uri)
}

func (r *Registry) loginableByProvider(provider string) (Loginable, error) {
	p, ok := r.GetProvider(provider)
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
	return DefaultRegistry.LoginByProvider(provider, username, password)
}

func (r *Registry) LoginByProvider(provider, username, password string) error {
	loginable, err := r.loginableByProvider(provider)
	if err != nil {
		return err
	}
	return loginable.Login(username, password)
}

func LogoutByProvider(provider string) error {
	return DefaultRegistry.LogoutByProvider(provider)
}

func (r *Registry) LogoutByProvider(provider string) error {
	loginable, err := r.loginableByProvider(provider)
	if err != nil {
		return err
	}
	return loginable.Logout()
}

func IsLoginByProvider(provider string) (bool, error) {
	return DefaultRegistry.IsLoginByProvider(provider)
}

func (r *Registry) IsLoginByProvider(provider string) (bool, error) {
	loginable, err := r.loginableByProvider(provider)
	if err != nil {
		return false, err
	}
	return loginable.IsLogin(), nil
}

func RefreshLoginByProvider(provider string) error {
	return DefaultRegistry.RefreshLoginByProvider(provider)
}

func (r *Registry) RefreshLoginByProvider(provider string) error {
	loginable, err := r.loginableByProvider(provider)
	if err != nil {
		return err
	}
	return loginable.RefreshLogin()
}

func QrLoginByProvider(provider string) (*QrLoginSession, error) {
	return DefaultRegistry.QrLoginByProvider(provider)
}

func (r *Registry) QrLoginByProvider(provider string) (*QrLoginSession, error) {
	loginable, err := r.loginableByProvider(provider)
	if err != nil {
		return nil, err
	}
	return loginable.QrLogin()
}

func QrLoginVerifyByProvider(provider string, qrlogin *QrLoginSession) (*QrLoginResult, error) {
	return DefaultRegistry.QrLoginVerifyByProvider(provider, qrlogin)
}

func (r *Registry) QrLoginVerifyByProvider(provider string, qrlogin *QrLoginSession) (*QrLoginResult, error) {
	loginable, err := r.loginableByProvider(provider)
	if err != nil {
		return nil, err
	}
	return loginable.QrLoginVerify(qrlogin)
}

func RestoreSessionByProvider(provider, session string) error {
	return DefaultRegistry.RestoreSessionByProvider(provider, session)
}

func (r *Registry) RestoreSessionByProvider(provider, session string) error {
	loginable, err := r.loginableByProvider(provider)
	if err != nil {
		return err
	}
	return loginable.RestoreSession(session)
}

func SaveSessionByProvider(provider string) (string, error) {
	return DefaultRegistry.SaveSessionByProvider(provider)
}

func (r *Registry) SaveSessionByProvider(provider string) (string, error) {
	loginable, err := r.loginableByProvider(provider)
	if err != nil {
		return "", err
	}
	return loginable.SaveSession(), nil
}
