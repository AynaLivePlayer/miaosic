package local

import (
	"os"
	"path"
	"strings"
	"sync"

	"github.com/AynaLivePlayer/miaosic"
)

type localPlaylist struct {
	name   string
	medias []localMedia
}

type localMedia struct {
	info        miaosic.MediaInfo
	quality     miaosic.Quality
	size        int64
	modTimeNano int64
	tagScanned  bool
	tagLoaded   bool
	coverLoaded bool
	search      string
}

func (l *localPlaylist) GetMediaInfo(meta miaosic.MetaData) (miaosic.MediaInfo, error) {
	for _, m := range l.medias {
		if m.info.Meta.Identifier == meta.Identifier {
			return m.info, nil
		}
	}
	return miaosic.MediaInfo{}, miaosic.ErrorInvalidMediaMeta
}

type Local struct {
	localDir             string
	playlists            map[string]*localPlaylist
	scanMode             TagScanMode
	cacheMode            LocalCacheMode
	mediaByID            map[string]*localMedia
	searchDocs           []localSearchDoc
	searchDocByID        map[string]int
	backgroundScanActive bool
	mu                   sync.RWMutex
}

func NewLocal(localdir string, options ...LocalOption) *Local {
	opts := localOptions{
		scanMode:  ScanTagOnStartup,
		cacheMode: UseLocalCache,
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		option.applyLocalOption(&opts)
	}
	l := &Local{
		localDir:      localdir,
		playlists:     make(map[string]*localPlaylist, 0),
		scanMode:      opts.scanMode,
		cacheMode:     opts.cacheMode,
		mediaByID:     make(map[string]*localMedia, 0),
		searchDocByID: make(map[string]int, 0),
	}
	if err := os.MkdirAll(localdir, 0755); err != nil {
		return l
	}
	for _, n := range getPlaylistNames(localdir) {
		playlist := &localPlaylist{name: n, medias: make([]localMedia, 0)}
		if readLocalPlaylist(localdir, playlist) == nil {
			l.playlists[playlist.name] = playlist
		}
	}
	l.rebuildIndexes()
	l.applyTagCache()
	switch opts.scanMode {
	case ScanTagOnStartup:
		l.scanTagsOnStartup()
	case ScanTagInBackground:
		l.scanTagsInBackground()
	default:
		// do nothing
	}
	return l
}

func (l *Local) metaToId(meta miaosic.MetaData) (playlist string) {
	return strings.Split(meta.Identifier, "/")[0]
}

func (l *Local) GetName() string {
	return "local"
}

func (l *Local) MatchMedia(uri string) (miaosic.MetaData, bool) {
	return miaosic.MetaData{}, false
}

func (l *Local) GetMediaInfo(meta miaosic.MetaData) (miaosic.MediaInfo, error) {
	if meta.Provider != l.GetName() {
		return miaosic.MediaInfo{}, miaosic.ErrorDifferentProvider
	}
	info, changed, err := l.loadMediaTag(meta.Identifier, true)
	if err == miaosic.ErrorInvalidMediaMeta {
		return info, err
	}
	if err == nil && changed {
		l.saveTagCache()
	}
	if err != nil {
		return info, nil
	}
	return info, nil
}

func (l *Local) GetMediaUrl(meta miaosic.MetaData, quality miaosic.Quality) ([]miaosic.MediaUrl, error) {
	info, err := l.GetMediaInfo(meta)
	if err != nil {
		return []miaosic.MediaUrl{}, err
	}
	return []miaosic.MediaUrl{{
		Url:     path.Join(l.localDir, info.Meta.Identifier),
		Quality: l.MapQuality(quality),
	}}, nil
}

func (l *Local) GetMediaLyric(meta miaosic.MetaData) ([]miaosic.Lyrics, error) {
	lyrics, err := readLyric(l.localDir, meta)
	if err == nil {
		return lyrics, nil
	}
	if len(lyrics) == 0 {
		return lyrics, miaosic.ErrorExternalApi
	}
	return []miaosic.Lyrics{}, miaosic.ErrorExternalApi
}

func (l *Local) Search(keyword string, page, size int) ([]miaosic.MediaInfo, error) {
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		return []miaosic.MediaInfo{}, nil
	}
	l.mu.RLock()
	docs := append([]localSearchDoc(nil), l.searchDocs...)
	l.mu.RUnlock()
	rankedMedias := rankLocalSearchDocs(keyword, docs)
	total := len(rankedMedias)
	startIdx := (page - 1) * size
	endIdx := page * size
	if startIdx >= total {
		return []miaosic.MediaInfo{}, nil
	}
	if endIdx >= total {
		endIdx = total
	}
	return rankedMedias[startIdx:endIdx], nil
}

func (l *Local) rebuildIndexes() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rebuildIndexesLocked()
}

func (l *Local) rebuildIndexesLocked() {
	l.mediaByID = make(map[string]*localMedia)
	l.searchDocs = l.searchDocs[:0]
	l.searchDocByID = make(map[string]int)
	for _, playlist := range l.playlists {
		for idx := range playlist.medias {
			media := &playlist.medias[idx]
			identifier := media.info.Meta.Identifier
			l.mediaByID[identifier] = media
			if media.search == "" {
				media.search = localSearchText(media.info)
			}
			l.searchDocs = append(l.searchDocs, localSearchDoc{
				info:   media.info,
				search: media.search,
			})
			l.searchDocByID[identifier] = len(l.searchDocs) - 1
		}
	}
}
