package tag

import (
	"fmt"
	"github.com/bogem/id3v2/v2"
	"io"
	"os"
	"strings"
)

func setID3v2Metadata(tag *id3v2.Tag, meta Metadata) {
	tag.DeleteAllFrames()
	// Use Unicode-safe default encoding for text frames like TIT2/TPE1/TALB.
	tag.SetDefaultEncoding(id3v2.EncodingUTF8)
	tag.SetTitle(meta.Title)
	tag.SetArtist(meta.Artist)
	tag.SetAlbum(meta.Album)
	for _, lyric := range meta.Lyrics {
		uslf := id3v2.UnsynchronisedLyricsFrame{
			Encoding: id3v2.EncodingUTF8,
			Language: normalizeID3Language(lyric.Lang),
			Lyrics:   lyric.Lyrics,
		}
		tag.AddUnsynchronisedLyricsFrame(uslf)
		tag.AddUserDefinedTextFrame(id3v2.UserDefinedTextFrame{
			Encoding:    id3v2.EncodingUTF8,
			Description: "lyrics",
			Value:       lyric.Lyrics,
		})
	}
	for _, pic := range meta.Pictures {
		picFrame := id3v2.PictureFrame{Encoding: id3v2.EncodingUTF8, MimeType: pic.Mimetype, PictureType: pic.Type, Description: pic.Description, Picture: pic.Data}
		tag.AddAttachedPicture(picFrame)
	}
}

func normalizeID3Language(lang string) string {
	l := strings.ToLower(strings.TrimSpace(lang))
	switch l {
	case "", "unk", "unknown":
		return "und"
	case "zh", "chi", "zho", "zh-cn", "zh-hans", "zh-hant":
		return "zho"
	case "en":
		return "eng"
	case "ja", "jp":
		return "jpn"
	case "ko":
		return "kor"
	}
	if len(l) >= 3 {
		c := l[:3]
		for i := 0; i < 3; i++ {
			if c[i] < 'a' || c[i] > 'z' {
				return "und"
			}
		}
		return c
	}
	return "und"
}

func WriteID3v2Tags(f *os.File, meta Metadata) error {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	tag, err := id3v2.ParseReader(f, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("error parsing mp3 file: %w", err)
	}
	setID3v2Metadata(tag, meta)
	return tag.Save()
}
