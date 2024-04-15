package local

import "github.com/AynaLivePlayer/miaosic"

func (l *Local) MatchPlaylist(uri string) (miaosic.MetaData, bool) {
	return miaosic.MetaData{
		Provider:   l.GetName(),
		Identifier: uri,
	}, true
}

func (l *Local) GetPlaylist(meta miaosic.MetaData) (*miaosic.Playlist, error) {
	playlist := &localPlaylist{name: meta.Identifier, medias: make([]localMedia, 0)}
	err := readLocalPlaylist(l.localDir, playlist)
	if err != nil {
		return nil, err

	}
	l.playlists[playlist.name] = playlist
	medias := make([]miaosic.MediaInfo, 0)
	for _, m := range playlist.medias {
		medias = append(medias, m.info)
	}
	return &miaosic.Playlist{
		Title:  meta.Identifier,
		Medias: medias,
		Meta:   meta,
	}, nil
}
