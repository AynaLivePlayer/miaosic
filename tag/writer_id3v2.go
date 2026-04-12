package tag

import (
	"fmt"
	"github.com/bogem/id3v2/v2"
	"io"
	"os"
)

func setID3v2Metadata(tag *id3v2.Tag, meta Metadata) {
	tag.DeleteAllFrames()
	tag.SetTitle(meta.Title)
	tag.SetArtist(meta.Artist)
	tag.SetAlbum(meta.Album)
	for _, lyric := range meta.Lyrics {
		lang := lyric.Lang
		if lang == "" {
			lang = "unk"
		}
		uslf := id3v2.UnsynchronisedLyricsFrame{Encoding: id3v2.EncodingUTF8, Language: lang[:min(3, len(lang))], Lyrics: lyric.Lyrics}
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
