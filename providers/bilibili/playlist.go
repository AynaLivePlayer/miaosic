package bilibili

import "miaosic"

func (n *Bilibili) MatchPlaylist(uri string) (miaosic.MetaData, bool) {
	return miaosic.MetaData{}, false
}

func (n *Bilibili) GetPlaylist(meta miaosic.MetaData) (*miaosic.Playlist, error) {
	return nil, miaosic.ErrNotImplemented
}
