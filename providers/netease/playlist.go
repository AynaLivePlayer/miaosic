package netease

import "miaosic"

func (n *Netease) MatchPlaylist(uri string) (miaosic.MetaData, bool) {
	return miaosic.MetaData{}, false
}

func (n *Netease) GetPlaylist(meta miaosic.MetaData) (*miaosic.Playlist, error) {
	return nil, miaosic.ErrNotImplemented
}
