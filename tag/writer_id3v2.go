package tag

import (
	"fmt"
	"github.com/bogem/id3v2/v2"
	"os"
)

func WriteID3v2Tags(f *os.File, meta Metadata) error {
	tag, err := id3v2.ParseReader(f, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("error parsing mp3 file: %w", err)
	}
	tag.SetTitle(meta.Title)
	tag.SetArtist(meta.Artist)
	tag.SetAlbum(meta.Album)
	for _, lyric := range meta.Lyrics {
		uslf := id3v2.UnsynchronisedLyricsFrame{Encoding: id3v2.EncodingUTF8, Language: lyric.Lang[:min(3, len(lyric.Lang))], Lyrics: lyric.Lyrics}
		tag.AddUnsynchronisedLyricsFrame(uslf)
	}
	for _, pic := range meta.Pictures {
		picFrame := id3v2.PictureFrame{Encoding: id3v2.EncodingUTF8, MimeType: pic.Mimetype, PictureType: pic.Type, Description: pic.Description, Picture: pic.Data}
		tag.AddAttachedPicture(picFrame)
	}
	return tag.Save()
}
