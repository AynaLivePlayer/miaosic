package tag

import (
	"github.com/bogem/id3v2/v2"
	"io"
)

func ReadID3v2Tags(r io.ReadSeeker) (Metadata, error) {
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
	tags, err := id3v2.ParseReader(r, id3v2.Options{Parse: true})
	if err != nil {
		return meta, err
	}
	if tags.Version() == 3 {
		meta.Format = FormatID3v2_3
	}
	if tags.Version() == 4 {
		meta.Format = FormatID3v2_4
	}
	meta.Title = tags.Title()
	meta.Artist = tags.Artist()
	meta.Album = tags.Album()
	meta.Lyrics = make([]Lyrics, 0)
	for _, frame := range tags.GetFrames("USLT") {
		lyricFrame, ok := frame.(id3v2.UnsynchronisedLyricsFrame)
		if !ok {
			continue
		}
		meta.Lyrics = append(meta.Lyrics, Lyrics{
			Lang:   lyricFrame.Language,
			Lyrics: lyricFrame.Lyrics,
		})
	}
	if len(meta.Lyrics) == 0 {
		for _, frame := range tags.GetFrames("TXXX") {
			userTextFrame, ok := frame.(id3v2.UserDefinedTextFrame)
			if !ok || userTextFrame.Value == "" {
				continue
			}
			switch userTextFrame.Description {
			case "lyrics", "USLT":
			default:
				continue
			}
			meta.Lyrics = append(meta.Lyrics, Lyrics{
				Lang:   "unk",
				Lyrics: userTextFrame.Value,
			})
		}
	}
	meta.Pictures = make([]Picture, 0)
	for _, frame := range tags.GetFrames("APIC") {
		pic, ok := frame.(id3v2.PictureFrame)
		if !ok {
			continue
		}
		meta.Pictures = append(meta.Pictures, Picture{
			Mimetype:    pic.MimeType,
			Type:        pic.PictureType,
			Description: pic.Description,
			Data:        pic.Picture,
		})
	}
	return meta, nil
}
