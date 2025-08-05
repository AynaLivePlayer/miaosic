package tag

type Picture struct {
	Mimetype    string
	Type        byte
	Description string
	Data        []byte
}

func (p Picture) TypeName() string {
	return pictureTypes[p.Type]
}

type Lyrics struct {
	Lang   string `json:"lang"`
	Lyrics string `json:"lyrics"`
}

type Metadata struct {
	Format   string `json:"format"`
	Mimetype string `json:"mimetype"`

	Title  string `json:"title"`
	Artist string `json:"artist"`
	Album  string `json:"album"`

	Lyrics   []Lyrics  `json:"lyrics"`
	Pictures []Picture `json:"pictures"`
}
