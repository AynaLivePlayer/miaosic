package local

import (
	"github.com/dhowden/tag"
	"github.com/sahilm/fuzzy"
	"miaosic"
	"os"
	"path"
	"path/filepath"
	"sort"
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
			if readMediaFile(localdir, &media) != nil {
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
	media.lyrics = []miaosic.Lyrics{miaosic.ParseLyrics("default", meta.Lyrics())}
	if meta.Picture() != nil {
		media.info.Cover.Data = meta.Picture().Data
	}
	data, err := os.ReadFile(filepath.Dir(p) + filepath.Base(p) + ".lrc")
	if err == nil && len(data) > 0 {
		media.lyrics = append(media.lyrics, miaosic.ParseLyrics("default", string(data)))
	}
	return nil
}

type mediaRanking struct {
	media *miaosic.MediaInfo
	score int
}

func rankMedia(keyword string, medias *[]miaosic.MediaInfo) []miaosic.MediaInfo {
	patterns := strings.Split(keyword, " ")
	data := make([]*mediaRanking, 0)

	for i, _ := range *medias {
		data = append(data, &mediaRanking{
			media: &(*medias)[i],
			score: 0,
		})
	}

	for _, pattern := range patterns {
		pattern = strings.ToLower(pattern)
		dataStr := make([]string, 0)
		for _, d := range data {
			dataStr = append(dataStr, strings.ToLower(d.media.Title))
		}
		for _, match := range fuzzy.Find(pattern, dataStr) {
			data[match.Index].score += match.Score
		}
		dataStr = make([]string, 0)
		for _, d := range data {
			dataStr = append(dataStr, strings.ToLower(d.media.Artist))
		}
		for _, match := range fuzzy.Find(pattern, dataStr) {
			data[match.Index].score += match.Score
		}
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].score > data[j].score
	})

	result := make([]miaosic.MediaInfo, 0)
	for _, d := range data {
		if d.score > 0 {
			result = append(result, *d.media)
		}
	}
	return result
}
