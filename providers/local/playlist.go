package local

import "github.com/AynaLivePlayer/miaosic"

func (l *Local) MatchPlaylist(uri string) (miaosic.MetaData, bool) {
	return miaosic.MetaData{}, false
}

func (l *Local) GetPlaylist(meta miaosic.MetaData) (*miaosic.Playlist, error) {
	return nil, miaosic.ErrNotImplemented
}
