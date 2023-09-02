package providers

import (
	"github.com/aynakeya/deepcolor/dphttp"
	"miaosic"
)

type DeepcolorProvider struct {
	InfoFunc     dphttp.ApiFunc[*miaosic.Media, *miaosic.Media]
	FileFunc     dphttp.ApiFunc[*miaosic.Media, *miaosic.Media]
	LyricFunc    dphttp.ApiFunc[*miaosic.Media, *miaosic.Media]
	PlaylistFunc dphttp.ApiFunc[*miaosic.Playlist, *miaosic.Playlist]
	SearchFunc   dphttp.ApiFuncResult[string, []*miaosic.Media]
}

func (d *DeepcolorProvider) UpdatePlaylist(playlist *miaosic.Playlist) error {
	if d.PlaylistFunc == nil {
		return miaosic.ErrNotImplemented
	}
	return d.PlaylistFunc(playlist, playlist)
}

func (d *DeepcolorProvider) Search(keyword string) ([]*miaosic.Media, error) {
	if d.SearchFunc == nil {
		return nil, miaosic.ErrNotImplemented
	}
	//result := make([]*miaosic.Media, 0)
	//err :=
	//fmt.Println(result)
	return d.SearchFunc(keyword)
}

func (d *DeepcolorProvider) UpdateMedia(media *miaosic.Media) error {
	if d.InfoFunc == nil {
		return miaosic.ErrNotImplemented
	}
	return d.InfoFunc(media, media)
}

func (d *DeepcolorProvider) UpdateMediaUrl(media *miaosic.Media) error {
	if d.FileFunc == nil {
		return miaosic.ErrNotImplemented
	}
	return d.FileFunc(media, media)
}

func (d *DeepcolorProvider) UpdateMediaLyric(media *miaosic.Media) error {
	media.Lyric = nil
	if d.LyricFunc == nil {
		// if no lyric func, return nil
		return nil
	}
	return d.LyricFunc(media, media)
}
