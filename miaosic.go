package miaosic

type Picture struct {
	Url  string
	Data []byte
}

func (p Picture) Exists() bool {
	return p.Url != "" || p.Data != nil
}

type MediaMeta struct {
	Provider   string
	Identifier string
}

func NewMediaMeta(provider, identifier string) MediaMeta {
	return MediaMeta{
		Provider:   provider,
		Identifier: identifier,
	}
}

func (m MediaMeta) ID() string {
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
	Meta   MediaMeta
}

//type Playlist struct {
//	Title  string
//	Medias []*Media
//	Meta   MediaMeta
//}

type MediaProvider interface {
	// GetName returns the name of the provider.
	GetName() string
	// Search returns a list of MediaMeta.
	Search(keyword string, page, size int) ([]MediaInfo, error)
	// MatchMedia returns a MediaMeta if the uri is matched, otherwise nil.
	MatchMedia(uri string) (MediaMeta, bool)
	GetMediaInfo(meta MediaMeta) (MediaInfo, error)
	GetMediaUrl(meta MediaMeta, quality Quality) ([]MediaUrl, error)
	GetMediaLyric(meta MediaMeta) ([]Lyrics, error)
	//// MatchPlaylist returns a MediaMeta if the uri is matched, otherwise nil.
	//MatchPlaylist(uri string) MediaMeta
	//GetPlaylist(meta MediaMeta) (*Playlist, error)
}

type Loginable interface {
	Login(username string, password string) error
	QrLogin() string
	QrLoginVerify() bool
	RestoreSession(session string) error
	SaveSession() string
}
