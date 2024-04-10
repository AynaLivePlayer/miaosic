package providers

import (
	"github.com/AynaLivePlayer/miaosic"
	"github.com/aynakeya/deepcolor/dphttp"
)

type FileApiParam struct {
	Meta    miaosic.MetaData
	Quality miaosic.Quality
}

type MediaSearchParam struct {
	Keyword  string
	Page     int
	PageSize int
}

type DeepcolorProvider struct {
	InfoApi  dphttp.ApiResultFunc[miaosic.MetaData, miaosic.MediaInfo]
	FileApi  dphttp.ApiResultFunc[FileApiParam, []miaosic.MediaUrl]
	LyricApi dphttp.ApiResultFunc[miaosic.MetaData, []miaosic.Lyrics]
	//PlaylistFunc dphttp.ApiFunc[*miaosic.Playlist, *miaosic.Playlist]
	SearchApi dphttp.ApiResultFunc[MediaSearchParam, []miaosic.MediaInfo]
}

func (p *DeepcolorProvider) Search(keyword string, page, size int) ([]miaosic.MediaInfo, error) {
	if p.SearchApi == nil {
		return nil, miaosic.ErrNotImplemented
	}
	return p.SearchApi(MediaSearchParam{Keyword: keyword, Page: page, PageSize: size})
}

func (p *DeepcolorProvider) GetMediaInfo(meta miaosic.MetaData) (miaosic.MediaInfo, error) {
	if p.InfoApi == nil {
		return miaosic.MediaInfo{}, miaosic.ErrNotImplemented
	}
	val, err := p.InfoApi(meta)
	val.Meta = meta
	return val, err
}

func (p *DeepcolorProvider) GetMediaUrl(meta miaosic.MetaData, quality miaosic.Quality) ([]miaosic.MediaUrl, error) {
	if p.FileApi == nil {
		return nil, miaosic.ErrNotImplemented
	}
	return p.FileApi(FileApiParam{Meta: meta, Quality: quality})
}

func (p *DeepcolorProvider) GetMediaLyric(meta miaosic.MetaData) ([]miaosic.Lyrics, error) {
	if p.LyricApi == nil {
		return nil, miaosic.ErrNotImplemented
	}
	return p.LyricApi(meta)
}

//func (d *DeepcolorProvider) UpdatePlaylist(playlist *miaosic.Playlist) error {
//	if d.PlaylistFunc == nil {
//		return miaosic.ErrNotImplemented
//	}
//	return d.PlaylistFunc(playlist, playlist)
//}
//
//func (d *DeepcolorProvider) Search(keyword string) ([]*miaosic.Media, error) {
//	if d.SearchFunc == nil {
//		return nil, miaosic.ErrNotImplemented
//	}
//	//result := make([]*miaosic.Media, 0)
//	//err :=
//	//fmt.Println(result)
//	return d.SearchFunc(keyword)
//}
//
//func (d *DeepcolorProvider) UpdateMedia(media *miaosic.Media) error {
//	if d.InfoFunc == nil {
//		return miaosic.ErrNotImplemented
//	}
//	return d.InfoFunc(media, media)
//}
//
//func (d *DeepcolorProvider) UpdateMediaUrl(media *miaosic.Media) error {
//	if d.FileFunc == nil {
//		return miaosic.ErrNotImplemented
//	}
//	return d.FileFunc(media, media)
//}
//
//func (d *DeepcolorProvider) UpdateMediaLyric(media *miaosic.Media) error {
//	media.Lyric = nil
//	if d.LyricFunc == nil {
//		// if no lyric func, return nil
//		return nil
//	}
//	return d.LyricFunc(media, media)
//}
