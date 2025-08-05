package tag

import (
	"github.com/dhowden/tag"
	"github.com/gabriel-vasile/mimetype"
	"io"
)

func Read(r io.ReadSeeker) (Metadata, error) {
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return Metadata{}, err
	}

	b := make([]byte, 512)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return Metadata{}, err
	}
	mimeType := mimetype.Detect(b).String()
	_, err = r.Seek(0, io.SeekStart)
	switch {
	case string(b[0:4]) == "fLaC":
		return ReadFLACTags(r, mimeType)
	//case string(b[0:4]) == "OggS":
	//	return ReadOGGTags(r)
	//case string(b[4:8]) == "ftyp":
	//	return ReadAtoms(r)
	case string(b[0:3]) == "ID3":
		return ReadID3v2Tags(r, mimeType)
		//case string(b[0:4]) == "DSD ":
		//	return ReadDSFTags(r)
	}
	return fallbackRead(r, mimeType)
}

func fallbackRead(r io.ReadSeeker, mime string) (Metadata, error) {
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
