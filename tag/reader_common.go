package tag

import (
	"io"

	"github.com/dhowden/tag"
)

func fallbackRead(r io.ReadSeeker) (Metadata, error) {
	mime, err := detectMime(r)
	if err != nil {
		return Metadata{}, err
	}
	if _, err = r.Seek(0, io.SeekStart); err != nil {
		return Metadata{}, err
	}
	meta := Metadata{
		Mimetype: mime,
	}
	m, err := tag.ReadFrom(r)
	if err != nil {
		return Metadata{}, err
	}
	meta.Format = string(m.Format())
	meta.Title = m.Title()
	meta.Artist = m.Artist()
	meta.Album = m.Album()
	meta.Lyrics = []Lyrics{}
	if m.Lyrics() != "" {
		meta.Lyrics = append(meta.Lyrics, Lyrics{
			Lang:   "unk",
			Lyrics: m.Lyrics(),
		})
	}
	meta.Pictures = []Picture{}
	if m.Picture() != nil {
		p := m.Picture()
		meta.Pictures = append(meta.Pictures, Picture{
			Mimetype:    p.MIMEType,
			Type:        PictureTypeFrontCover,
			Description: p.Description,
			Data:        p.Data,
		})
	}
	return meta, nil
}
