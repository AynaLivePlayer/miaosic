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
	l.applyTagCacheToPlaylist(playlist)
	l.applyExistingMediaToPlaylist(playlist)
	if l.scanMode == ScanTagOnStartup {
		l.scanPlaylistTags(playlist)
	}
	l.mu.Lock()
	l.playlists[playlist.name] = playlist
	l.rebuildIndexesLocked()
	l.mu.Unlock()

	result := localPlaylistToPlaylist(meta, playlist)
	if l.scanMode == ScanTagOnStartup {
		l.saveTagCache()
	}
	if l.scanMode == ScanTagInBackground {
		l.scanTagsInBackground()
	}
	return result, nil
}

func (l *Local) applyExistingMediaToPlaylist(playlist *localPlaylist) {
	l.mu.RLock()
	existing := l.playlists[playlist.name]
	if existing == nil {
		l.mu.RUnlock()
		return
	}
	existingMedias := make(map[string]localMedia, len(existing.medias))
	for _, media := range existing.medias {
		existingMedias[media.info.Meta.Identifier] = media
	}
	l.mu.RUnlock()

	for idx := range playlist.medias {
		media := &playlist.medias[idx]
		existingMedia, ok := existingMedias[media.info.Meta.Identifier]
		if !ok || existingMedia.size != media.size || existingMedia.modTimeNano != media.modTimeNano {
			continue
		}
		*media = existingMedia
	}
}

func localPlaylistToPlaylist(meta miaosic.MetaData, playlist *localPlaylist) *miaosic.Playlist {
	medias := make([]miaosic.MediaInfo, 0)
	for _, m := range playlist.medias {
		medias = append(medias, m.info)
	}
	return &miaosic.Playlist{
		Title:  meta.Identifier,
		Medias: medias,
		Meta:   meta,
	}
}
