package netease

import (
	"github.com/AynaLivePlayer/miaosic"
	neteaseApi "github.com/XiaoMengXinX/Music163Api-Go/api"
	"github.com/spf13/cast"
	"strconv"
)

func (n *Netease) MatchPlaylist(uri string) (miaosic.MetaData, bool) {
	var id string
	id = n.PlaylistRegex0.FindString(uri)
	if id != "" {
		return miaosic.MetaData{
			Provider:   n.GetName(),
			Identifier: id,
		}, true
	}
	id = n.PlaylistRegex1.FindString(uri)
	if id != "" {
		return miaosic.MetaData{
			Provider:   n.GetName(),
			Identifier: id[12:],
		}, true
	}
	return miaosic.MetaData{}, false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (n *Netease) GetPlaylist(meta miaosic.MetaData) (*miaosic.Playlist, error) {
	if meta.Provider != n.GetName() {
		return nil, miaosic.ErrorInvalidMediaMeta
	}

	result, err := neteaseApi.GetPlaylistDetail(
		n.ReqData, cast.ToInt(meta.Identifier))
	if err != nil || result.Code != 200 {
		return nil, miaosic.ErrorExternalApi
	}
	cnt := len(result.Playlist.TrackIds)
	if cnt == 0 {
		return nil, miaosic.ErrorExternalApi
	}

	playlist := &miaosic.Playlist{
		Meta:   meta,
		Medias: make([]miaosic.MediaInfo, 0),
		Title:  result.Playlist.Name,
	}

	ids := make([]int, len(result.Playlist.TrackIds))
	for i := 0; i < cnt; i++ {
		ids[i] = result.Playlist.TrackIds[i].Id
	}
	for index := 0; index < len(ids); index += 1000 {
		result2, err := neteaseApi.GetSongDetail(
			n.ReqData,
			ids[index:min(index+1000, len(ids))])
		if err != nil || result2.Code != 200 {
			break
		}
		cnt = len(result2.Songs)
		if cnt == 0 {
			break
		}
		for i := 0; i < cnt; i++ {
			playlist.Medias = append(playlist.Medias, miaosic.MediaInfo{
				Title:  result2.Songs[i].Name,
				Artist: _neteaseGetArtistNames(result2.Songs[i]),
				Cover:  miaosic.Picture{Url: result2.Songs[i].Al.PicUrl},
				Album:  result2.Songs[i].Al.Name,
				Meta: miaosic.MetaData{
					Provider:   n.GetName(),
					Identifier: strconv.Itoa(result2.Songs[i].Id),
				},
			})
		}
	}
	if len(playlist.Medias) == 0 {
		return nil, miaosic.ErrorExternalApi
	}
	return playlist, nil
}
