package local

import (
	"github.com/AynaLivePlayer/miaosic"
	"os"
	"path"
	"strings"
)

type localPlaylist struct {
	name   string
	medias []localMedia
}

type localMedia struct {
	info    miaosic.MediaInfo
	quality miaosic.Quality
	lyrics  []miaosic.Lyrics
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
	localDir  string
	playlists map[string]*localPlaylist
}

func NewLocal(localdir string) *Local {
	l := &Local{localDir: localdir, playlists: make(map[string]*localPlaylist, 0)}
	if err := os.MkdirAll(localdir, 0755); err != nil {
		return l
	}
	for _, n := range getPlaylistNames(localdir) {
		playlist := &localPlaylist{name: n, medias: make([]localMedia, 0)}
		if readLocalPlaylist(localdir, playlist) == nil {
			l.playlists[playlist.name] = playlist
		}
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
	playlist, ok := l.playlists[l.metaToId(meta)]
	if !ok {
		return miaosic.MediaInfo{}, miaosic.ErrorInvalidMediaMeta
	}
	for _, m := range playlist.medias {
		if m.info.Meta.Identifier == meta.Identifier {
			newM := m
			_ = readMediaFile(l.localDir, &newM)
			return newM.info, nil
		}
	}
	return miaosic.MediaInfo{}, miaosic.ErrorInvalidMediaMeta
}

func (l *Local) GetMediaUrl(meta miaosic.MetaData, quality miaosic.Quality) ([]miaosic.MediaUrl, error) {
	info, err := l.GetMediaInfo(meta)
	if err != nil {
		return []miaosic.MediaUrl{}, err
	}
	return []miaosic.MediaUrl{{
		Url:     path.Join(l.localDir, info.Meta.Identifier),
		Quality: miaosic.QualityUnk,
	}}, nil
}

func (l *Local) GetMediaLyric(meta miaosic.MetaData) ([]miaosic.Lyrics, error) {
	playlist, ok := l.playlists[l.metaToId(meta)]
	if !ok {
		return []miaosic.Lyrics{}, miaosic.ErrorInvalidMediaMeta
	}
	for _, m := range playlist.medias {
		if m.info.Meta.Identifier == meta.Identifier {
			newM := m
			_ = readMediaFile(l.localDir, &newM)
			return newM.lyrics, nil
		}
	}
	return []miaosic.Lyrics{}, miaosic.ErrorExternalApi
}

func (l *Local) Search(keyword string, page, size int) ([]miaosic.MediaInfo, error) {
	allMedias := make([]miaosic.MediaInfo, 0)
	for _, p := range l.playlists {
		for _, m := range p.medias {
			allMedias = append(allMedias, m.info)
		}
	}
	rankedMedias := rankMedia(keyword, &allMedias)
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
