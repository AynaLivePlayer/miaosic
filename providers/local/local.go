package local

import (
	"miaosic"
	"os"
	"path"
)

type Local struct {
	localDir  string
	playlists map[string]*miaosic.Playlist
}

func NewLocal(localdir string) *Local {
	l := &Local{localDir: localdir, playlists: make(map[string]*miaosic.Playlist, 0)}
	if err := os.MkdirAll(localdir, 0755); err != nil {
		return l
	}
	for _, n := range getPlaylistNames(localdir) {
		playlist := &miaosic.Playlist{Meta: miaosic.MediaMeta{Provider: n}}
		if readLocalPlaylist(localdir, playlist) != nil {
			l.playlists[playlist.Title] = playlist
		}
	}
	return l
}

func (l *Local) GetName() string {
	return "local"
}

func (l *Local) MatchMedia(uri string) *miaosic.Media {
	return nil
}

func (l *Local) MatchPlaylist(uri string) *miaosic.Playlist {
	return nil
}

func (l *Local) Search(keyword string) ([]*miaosic.Media, error) {
	allMedias := make([]*miaosic.Media, 0)
	for _, p := range l.playlists {
		for _, m := range p.Medias {
			allMedias = append(allMedias, m)
		}
	}
	return RankMedia(keyword, allMedias), nil
}

func (l *Local) UpdatePlaylist(playlist *miaosic.Playlist) error {
	err := readLocalPlaylist(l.localDir, playlist)
	if err != nil {
		return err
	}
	l.playlists[playlist.Meta.Identifier] = playlist
	return nil
}

func (l *Local) UpdateMedia(media *miaosic.Media) error {
	mediaPath := path.Join(l.localDir, media.Meta.Identifier)
	_, err := os.Stat(mediaPath)
	if err != nil {
		return err
	}
	return readMediaFile(l.localDir, media)
}

func (l *Local) UpdateMediaUrl(media *miaosic.Media) error {
	mediaPath := path.Join(l.localDir, media.Meta.Identifier)
	_, err := os.Stat(mediaPath)
	if err != nil {
		return err
	}
	media.Url = mediaPath
	return nil
}

func (l *Local) UpdateMediaLyric(media *miaosic.Media) error {
	return nil
}

//
//func (l *Local) Search(keyword string) ([]*model.Media, error) {
//	allMedias := make([]*model.Media, 0)
//	for _, p := range l.Playlists {
//		for _, m := range p.Medias {
//			allMedias = append(allMedias, m)
//		}
//	}
//	MediaSort(keyword, allMedias)
//	c := util.Min(len(allMedias), 32)
//	medias := make([]*model.Media, c)
//	for i := 0; i < c; i++ {
//		medias[i] = allMedias[i].Copy()
//	}
//	return medias, nil
//}
//
//func (l *Local) SearchV1(keyword string) ([]*model.Media, error) {
//	result := make([]struct {
//		M *model.Media
//		N int
//	}, 0)
//	keywords := strings.Split(keyword, " ")
//	for _, p := range l.Playlists {
//		for _, m := range p.Medias {
//			title := strings.ToLower(m.Title)
//			artist := strings.ToLower(m.Artist)
//			n := 0
//			for _, k := range keywords {
//				kw := strings.ToLower(k)
//				if strings.Contains(title, kw) || strings.Contains(artist, kw) {
//					n++
//				}
//				if kw == title {
//					n += 3
//				}
//			}
//			if n > 0 {
//				result = append(result, struct {
//					M *model.Media
//					N int
//				}{M: m, N: n})
//			}
//		}
//	}
//	sort.Slice(result, func(i, j int) bool {
//		return result[i].N > result[j].N
//	})
//	medias := make([]*model.Media, len(result))
//	for i, r := range result {
//		medias[i] = r.M.Copy()
//	}
//	return medias, nil
//}
