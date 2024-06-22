package kugou

import "github.com/AynaLivePlayer/miaosic"

func (k *Kugou) MatchPlaylist(uri string) (miaosic.MetaData, bool) {
	return miaosic.MetaData{}, false
}

func (k *Kugou) GetPlaylist(meta miaosic.MetaData) (*miaosic.Playlist, error) {
	return nil, miaosic.ErrNotImplemented
}
