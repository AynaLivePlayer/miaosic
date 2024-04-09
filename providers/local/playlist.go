package local

import "github.com/AynaLivePlayer/miaosic"

func (l *localPlaylist) MatchPlaylist(uri string) (miaosic.MetaData, bool) {
	return miaosic.MetaData{}, false
}

func (l *localPlaylist) GetPlaylist(meta miaosic.MetaData) (*miaosic.Playlist, error) {
	return nil, miaosic.ErrNotImplemented
}
