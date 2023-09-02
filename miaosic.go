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

func (m MediaMeta) ID() string {
	return m.Provider + "_" + m.Identifier
}

type Media struct {
	Title  string
	Artist string
	Cover  Picture
	Album  string
	Lyric  string
	Url    string
	Header map[string]string
	Meta   MediaMeta
}

type Playlist struct {
	Title  string
	Medias []*Media
	Meta   MediaMeta
}

type MediaProvider interface {
	GetName() string
	// MatchMedia returns a Media if the uri is matched, otherwise nil.
	MatchMedia(uri string) *Media
	// MatchPlaylist returns a Playlist if the uri is matched, otherwise nil.
	MatchPlaylist(uri string) *Playlist
	Search(keyword string) ([]*Media, error)
	UpdatePlaylist(playlist *Playlist) error
	UpdateMedia(media *Media) error
	UpdateMediaUrl(media *Media) error
	UpdateMediaLyric(media *Media) error
}
