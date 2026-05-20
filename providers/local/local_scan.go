package local

import "github.com/AynaLivePlayer/miaosic"

type LocalOption interface {
	applyLocalOption(*localOptions)
}

type localOptions struct {
	scanMode  TagScanMode
	cacheMode LocalCacheMode
}

type TagScanMode int

const (
	DontScanTag TagScanMode = iota
	ScanTagOnStartup
	ScanTagInBackground
)

func (m TagScanMode) applyLocalOption(options *localOptions) {
	options.scanMode = m
}

func (l *Local) scanTagsOnStartup() {
	for _, playlist := range l.playlists {
		l.scanPlaylistTags(playlist)
	}
	l.rebuildIndexes()
	l.saveTagCache()
}

func (l *Local) scanPlaylistTags(playlist *localPlaylist) {
	for idx := range playlist.medias {
		media := &playlist.medias[idx]
		if media.tagScanned {
			continue
		}
		if err := readMediaFileTag(l.localDir, media, false); err == nil {
			media.search = localSearchText(media.info)
			media.tagLoaded = true
		}
		media.tagScanned = true
	}
}

func (l *Local) scanTagsInBackground() {
	l.mu.Lock()
	if l.backgroundScanActive {
		l.mu.Unlock()
		return
	}
	l.backgroundScanActive = true
	l.mu.Unlock()

	go func() {
		defer func() {
			l.mu.Lock()
			l.backgroundScanActive = false
			l.mu.Unlock()
			if len(l.unloadedMediaIdentifiers()) > 0 {
				l.scanTagsInBackground()
			}
		}()

		for {
			identifiers := l.unloadedMediaIdentifiers()
			if len(identifiers) == 0 {
				return
			}
			changed := false
			for _, identifier := range identifiers {
				if _, mediaChanged, err := l.loadMediaTag(identifier, false); err == nil {
					changed = changed || mediaChanged
				}
			}
			if changed {
				l.saveTagCache()
			}
		}
	}()
}

func (l *Local) unloadedMediaIdentifiers() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	identifiers := make([]string, 0)
	for _, playlist := range l.playlists {
		for idx := range playlist.medias {
			media := &playlist.medias[idx]
			if !media.tagScanned {
				identifiers = append(identifiers, media.info.Meta.Identifier)
			}
		}
	}
	return identifiers
}

func (l *Local) loadMediaTag(identifier string, includeCover bool) (miaosic.MediaInfo, bool, error) {
	l.mu.RLock()
	media, ok := l.mediaByID[identifier]
	if ok {
		info := media.info
		if media.tagLoaded && (!includeCover || media.coverLoaded) {
			l.mu.RUnlock()
			return info, false, nil
		}
		if media.tagScanned && !includeCover {
			l.mu.RUnlock()
			return info, false, nil
		}
		if includeCover && media.coverLoaded {
			l.mu.RUnlock()
			return info, false, nil
		}
	}
	if !ok {
		l.mu.RUnlock()
		return miaosic.MediaInfo{}, false, miaosic.ErrorInvalidMediaMeta
	}
	info := media.info
	l.mu.RUnlock()

	loadedMedia := localMedia{info: info}
	err := readMediaFileTag(l.localDir, &loadedMedia, includeCover)

	l.mu.Lock()
	defer l.mu.Unlock()
	media, ok = l.mediaByID[identifier]
	if !ok {
		return miaosic.MediaInfo{}, false, miaosic.ErrorInvalidMediaMeta
	}
	if media.tagLoaded && (!includeCover || media.coverLoaded) {
		return media.info, false, err
	}
	if media.tagScanned && !includeCover {
		return media.info, false, err
	}
	if includeCover && media.coverLoaded {
		return media.info, false, err
	}
	if err == nil {
		l.updateMediaTagLocked(media, loadedMedia, includeCover)
		return media.info, true, nil
	}
	media.tagScanned = true
	if includeCover {
		media.coverLoaded = true
	}
	return media.info, false, err
}

func (l *Local) updateMediaTagLocked(media *localMedia, loadedMedia localMedia, includeCover bool) {
	media.info = loadedMedia.info
	media.size = loadedMedia.size
	media.modTimeNano = loadedMedia.modTimeNano
	media.search = localSearchText(loadedMedia.info)
	media.tagScanned = true
	media.tagLoaded = true
	if includeCover {
		media.coverLoaded = true
	}

	if idx, ok := l.searchDocByID[loadedMedia.info.Meta.Identifier]; ok {
		if idx >= 0 && idx < len(l.searchDocs) {
			l.searchDocs[idx] = localSearchDoc{
				info:   loadedMedia.info,
				search: media.search,
			}
			return
		}
	}
	l.rebuildIndexesLocked()
}
