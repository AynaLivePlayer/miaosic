package local

import (
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/dhowden/tag"
	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var detector = chardet.NewTextDetector()

func getEncoding(name string) encoding.Encoding {
	switch name {
	case "UTF-8":
		return unicode.UTF8
	case "ISO-8859-1":
		return charmap.ISO8859_1
	case "Windows-1252":
		return charmap.Windows1252
	case "GBK":
		return simplifiedchinese.GBK
	case "GB-18030":
		return simplifiedchinese.GB18030
	case "Big5":
		return traditionalchinese.Big5
	case "Shift_JIS":
		return japanese.ShiftJIS
	case "EUC-KR":
		return korean.EUCKR
	default:
		return unicode.UTF8
	}
}

func decodeBytes(data []byte, enc encoding.Encoding) (string, error) {
	// Create a transformer that will decode the bytes
	transformer := enc.NewDecoder()

	// Transform the bytes into a string
	result, _, err := transform.String(transformer, string(data))
	if err != nil {
		return "", err
	}

	return result, nil
}

func getPlaylistNames(localdir string) []string {
	names := make([]string, 0)
	items, _ := os.ReadDir(localdir)
	for _, item := range items {
		if item.IsDir() {
			names = append(names, item.Name())
		}
	}
	return names
}

// readLocalPlaylist read files under a directory
// and return a _LocalPlaylist object.
// This function assume this directory exists
func readLocalPlaylist(localdir string, playlist *localPlaylist) error {
	playlist.medias = make([]localMedia, 0)
	fullPath := filepath.Join(localdir, playlist.name)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return err
	}
	items, _ := os.ReadDir(fullPath)
	for _, item := range items {
		// if item is a file, read file
		if !item.IsDir() {
			fn := item.Name()
			media := localMedia{
				info: miaosic.MediaInfo{
					Meta: miaosic.MetaData{
						Provider:   "local",
						Identifier: path.Join(playlist.name, fn),
					},
				},
			}
			if err := readMediaFileInfo(localdir, &media); err != nil {
				continue
			}
			playlist.medias = append(playlist.medias, media)
		}
	}
	return nil
}

func _getOrDefault(s string, def string) string {
	if s == "" {
		return def
	}
	return s
}

func readMediaFile(localdir string, media *localMedia) error {
	p := path.Join(localdir, media.info.Meta.Identifier)
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	meta, err := tag.ReadFrom(f)
	if err != nil {
		return err
	}
	media.info.Title = _getOrDefault(meta.Title(), filepath.Base(p))
	media.info.Artist = _getOrDefault(meta.Artist(), "Unknown")
	media.info.Album = _getOrDefault(meta.Album(), "Unknown")
	if meta.Picture() != nil {
		media.info.Cover.Data = meta.Picture().Data
	}
	return nil
}

func readMediaFileInfo(localdir string, media *localMedia) error {
	p := path.Join(localdir, media.info.Meta.Identifier)
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	meta, err := tag.ReadFrom(f)
	if err != nil {
		return err
	}
	media.info.Title = _getOrDefault(meta.Title(), filepath.Base(p))
	media.info.Artist = _getOrDefault(meta.Artist(), "Unknown")
	media.info.Album = _getOrDefault(meta.Album(), "Unknown")
	return nil
}

func readLyric(localdir string, meta miaosic.MetaData) ([]miaosic.Lyrics, error) {
	lyrics := make([]miaosic.Lyrics, 0)
	p := path.Join(localdir, meta.Identifier)

	data, err := os.ReadFile(path.Join(filepath.Dir(p), strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))+".lrc"))
	if err == nil && len(data) > 0 {
		detectedChar, err := detector.DetectBest(data)
		if err != nil {
			detectedChar.Charset = "UTF-8"
		}
		datastr, _ := decodeBytes(data, getEncoding(detectedChar.Charset))
		if datastr != "" {
			lyrics = append(lyrics, miaosic.ParseLyrics("default", datastr))
		}
	}
	f, err := os.Open(p)
	defer f.Close()
	if err == nil {
		mMeta, err := tag.ReadFrom(f)
		if err == nil {
			name := "builtin"
			if len(lyrics) == 0 {
				name = "default"
			}
			lyrics = append(lyrics, miaosic.ParseLyrics(name, mMeta.Lyrics()))
		}
	}
	//fmt.Println(lyrics)
	return lyrics, nil
}
