package local

import (
	"os"
	"path"
	"path/filepath"
	"strings"

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
	items, err := os.ReadDir(fullPath)
	if err != nil {
		return err
	}
	for _, item := range items {
		// if item is a file, read file
		if !item.IsDir() {
			fn := item.Name()
			if !isSupportedLocalAudioFile(fn) {
				continue
			}
			info, err := item.Info()
			if err != nil {
				continue
			}
			media := localMedia{
				size:        info.Size(),
				modTimeNano: info.ModTime().UnixNano(),
				info: miaosic.MediaInfo{
					Title:  fn,
					Artist: "Unknown",
					Album:  "Unknown",
					Meta: miaosic.MetaData{
						Provider:   "local",
						Identifier: path.Join(playlist.name, fn),
					},
				},
			}
			media.search = localSearchText(media.info)
			playlist.medias = append(playlist.medias, media)
		}
	}
	return nil
}

func isSupportedLocalAudioFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".mp3", ".flac", ".wav", ".ogg", ".opus", ".m4a", ".mp4", ".aac", ".alac":
		return true
	default:
		return false
	}
}

func _getOrDefault(s string, def string) string {
	if s == "" {
		return def
	}
	return s
}

func readMediaFileTag(localdir string, media *localMedia, includeCover bool) error {
	p := path.Join(localdir, media.info.Meta.Identifier)
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	media.size = info.Size()
	media.modTimeNano = info.ModTime().UnixNano()
	meta, err := tag.ReadFrom(f)
	if err != nil {
		return err
	}
	media.info.Title = _getOrDefault(meta.Title(), filepath.Base(p))
	media.info.Artist = _getOrDefault(meta.Artist(), "Unknown")
	media.info.Album = _getOrDefault(meta.Album(), "Unknown")
	if includeCover && meta.Picture() != nil {
		media.info.Cover.Data = meta.Picture().Data
	}
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
	if err == nil {
		defer f.Close()
		mMeta, err := tag.ReadFrom(f)
		if err == nil {
			name := "builtin"
			if len(lyrics) == 0 {
				name = "default"
			}
			lyrics = append(lyrics, miaosic.ParseLyrics(name, mMeta.Lyrics()))
		}
	}
	return lyrics, nil
}
