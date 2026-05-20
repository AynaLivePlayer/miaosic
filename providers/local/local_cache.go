package local

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const localCacheFileName = ".miaosic-local-cache.json"

type LocalCacheMode int

const (
	DontUseLocalCache LocalCacheMode = iota
	UseLocalCache
)

func (m LocalCacheMode) applyLocalOption(options *localOptions) {
	options.cacheMode = m
}

type localTagCache struct {
	Version int                           `json:"version"`
	Medias  map[string]localTagCacheMedia `json:"medias"`
}

type localTagCacheMedia struct {
	Size        int64    `json:"size"`
	ModTimeNano int64    `json:"mod_time_nano"`
	Title       string   `json:"title"`
	Artist      string   `json:"artist"`
	Artists     []string `json:"artists,omitempty"`
	Album       string   `json:"album"`
}

func (l *Local) cacheEnabled() bool {
	return l.cacheMode == UseLocalCache
}

func (l *Local) cachePath() string {
	return filepath.Join(l.localDir, localCacheFileName)
}

func (l *Local) applyTagCache() {
	if !l.cacheEnabled() {
		return
	}
	cache, err := readLocalTagCache(l.cachePath())
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	for _, playlist := range l.playlists {
		applyLocalTagCacheToPlaylist(playlist, cache)
	}
	l.rebuildIndexesLocked()
}

func (l *Local) applyTagCacheToPlaylist(playlist *localPlaylist) bool {
	if !l.cacheEnabled() {
		return false
	}
	cache, err := readLocalTagCache(l.cachePath())
	if err != nil {
		return false
	}
	return applyLocalTagCacheToPlaylist(playlist, cache)
}

func applyLocalTagCacheToPlaylist(playlist *localPlaylist, cache localTagCache) bool {
	changed := false
	for idx := range playlist.medias {
		media := &playlist.medias[idx]
		cached, ok := cache.Medias[media.info.Meta.Identifier]
		if !ok || cached.Size != media.size || cached.ModTimeNano != media.modTimeNano {
			continue
		}
		media.info.Title = cached.Title
		media.info.Artist = cached.Artist
		media.info.Artists = append([]string(nil), cached.Artists...)
		media.info.Album = cached.Album
		media.search = localSearchText(media.info)
		media.tagScanned = true
		media.tagLoaded = true
		media.coverLoaded = false
		changed = true
	}
	return changed
}

func (l *Local) saveTagCache() {
	if !l.cacheEnabled() {
		return
	}
	cache := localTagCache{
		Version: 1,
		Medias:  make(map[string]localTagCacheMedia),
	}

	l.mu.RLock()
	for _, playlist := range l.playlists {
		for idx := range playlist.medias {
			media := &playlist.medias[idx]
			if !media.tagLoaded {
				continue
			}
			cache.Medias[media.info.Meta.Identifier] = localTagCacheMedia{
				Size:        media.size,
				ModTimeNano: media.modTimeNano,
				Title:       media.info.Title,
				Artist:      media.info.Artist,
				Artists:     append([]string(nil), media.info.Artists...),
				Album:       media.info.Album,
			}
		}
	}
	l.mu.RUnlock()

	_ = writeLocalTagCache(l.cachePath(), cache)
}

func readLocalTagCache(filename string) (localTagCache, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return localTagCache{}, err
	}
	var cache localTagCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return localTagCache{}, err
	}
	if cache.Medias == nil {
		cache.Medias = make(map[string]localTagCacheMedia)
	}
	return cache, nil
}

func writeLocalTagCache(filename string, cache localTagCache) error {
	if cache.Medias == nil {
		cache.Medias = make(map[string]localTagCacheMedia)
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(filepath.Dir(filename), ".miaosic-local-cache-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, filename); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}
