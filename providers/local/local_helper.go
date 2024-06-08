package local

import (
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/dhowden/tag"
	"os"
	"path"
	"path/filepath"
	"strings"
)

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
		lyrics = append(lyrics, miaosic.ParseLyrics("default", string(data)))
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
	fmt.Println(lyrics)
	return lyrics, nil
}
