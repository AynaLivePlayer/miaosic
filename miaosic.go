package miaosic

type Picture struct {
	Url  string
	Data []byte
}

func (p Picture) Exists() bool {
	return p.Url != "" || p.Data != nil
}

type MetaData struct {
	Provider   string
	Identifier string
}

func NewMetaData(provider, identifier string) MetaData {
	return MetaData{
		Provider:   provider,
		Identifier: identifier,
	}
}

func (m MetaData) ID() string {
	return m.Provider + "_" + m.Identifier
}

type Quality string

const (
	QualityAny  Quality = ""
	QualityUnk  Quality = "unknown"
	Quality128k Quality = "128k"
	Quality192k Quality = "192k"
	Quality256k Quality = "256k"
	Quality320k Quality = "320k"
)

type MediaUrl struct {
	Url     string
	Quality Quality
	Header  map[string]string
}

func NewMediaUrl(url string, quality Quality) MediaUrl {
	return MediaUrl{
		Url:     url,
		Quality: quality,
		Header:  make(map[string]string),
	}
}

type MediaInfo struct {
	Title  string
	Artist string
	Cover  Picture
	Album  string
	Meta   MetaData
}

type Playlist struct {
	Title  string
	Medias []MediaInfo
	Meta   MetaData
}

func (p *Playlist) DisplayName() string {
	if p.Title != "" {
		return p.Title
	}
	return p.Meta.ID()
}

func (p *Playlist) Copy() Playlist {
	medias := make([]MediaInfo, len(p.Medias))
	copy(medias, p.Medias)
	return Playlist{
		Title:  p.Title,
		Medias: medias,
		Meta:   p.Meta,
	}
}

type MediaProvider interface {
	// GetName returns the name of the provider.
	GetName() string

	// Search returns a list of MetaData.
	Search(keyword string, page, size int) ([]MediaInfo, error)

	// ===== Media related =====

	// MatchMedia returns a MetaData if the uri is matched, otherwise nil.
	MatchMedia(uri string) (MetaData, bool)
	GetMediaInfo(meta MetaData) (MediaInfo, error)
	GetMediaUrl(meta MetaData, quality Quality) ([]MediaUrl, error)
	GetMediaLyric(meta MetaData) ([]Lyrics, error)

	// ===== Playlist related =====

	// MatchPlaylist returns a MetaData if the uri is matched, otherwise nil.
	MatchPlaylist(uri string) (MetaData, bool)
	// GetPlaylist returns a Playlist, it fetches all data, so it might be slow.
	GetPlaylist(meta MetaData) (*Playlist, error)
}

type QrLoginSession struct {
	Url string
	Key string
}

type QrLoginResult struct {
	Success bool
	Message string
}

type Loginable interface {
	Login(username string, password string) error
	Logout() error
	QrLogin() (*QrLoginSession, error)
	QrLoginVerify(qrlogin *QrLoginSession) (*QrLoginResult, error)
	RestoreSession(session string) error
	SaveSession() string
}
