package kuwo

import "github.com/AynaLivePlayer/miaosic"

func (n *Kuwo) MatchPlaylist(uri string) (miaosic.MetaData, bool) {
	return miaosic.MetaData{}, false
}

func (n *Kuwo) GetPlaylist(meta miaosic.MetaData) (*miaosic.Playlist, error) {
	return nil, miaosic.ErrNotImplemented
}

//func (k *Kuwo) MatchPlaylist(uri string) *miaosic.Playlist {
//	var id string
//	id = k.PlaylistRegex0.FindString(uri)
//	if id != "" {
//		return &miaosic.Playlist{
//			Meta: miaosic.MetaData{k.GetName(), id},
//		}
//	}
//	id = k.PlaylistRegex1.FindString(uri)
//	if id != "" {
//		return &miaosic.Playlist{
//			Meta: miaosic.MetaData{k.GetName(), id[9:]},
//		}
//	}
//	return nil
//}

//func (k *Kuwo) playlistApi(src *miaosic.Playlist, dst *miaosic.Playlist) error {
//	dst.Medias = make([]*miaosic.Media, 0)
//	api := deepcolor.CreateChainApiFunc(
//		k.requester,
//		func(page int) (*dphttp.Request, error) {
//			return deepcolor.NewGetRequestWithQuery(
//				"http://www.kuwo.cn/api/www/playlist/playListInfo",
//				[]string{"pid", "pn", "rn"}, k.header)([]string{src.Meta.Identifier, cast.ToString(page), "100"})
//		},
//		deepcolor.ParserGJson,
//		func(resp *gjson.Result, playlist *miaosic.Playlist) error {
//			resp.Get("data.musicList").ForEach(func(key, value gjson.Result) bool {
//				playlist.Medias = append(
//					playlist.Medias,
//					&miaosic.Media{
//						Title:  html.UnescapeString(value.Get("name").String()),
//						Artist: value.Get("artist").String(),
//						Cover:  miaosic.Picture{Url: value.Get("pic").String()},
//						Album:  value.Get("album").String(),
//						Meta: miaosic.MetaData{
//							Provider:   k.GetName(),
//							Identifier: value.Get("rid").String(),
//						},
//					})
//				return true
//			})
//			return nil
//		},
//		func(page int, resp *gjson.Result, playlist *miaosic.Playlist) (int, bool) {
//			if resp.Get("code").String() != "200" {
//				return page, false
//			}
//			cnt := int(resp.Get("data.total").Int())
//			if cnt <= page*100 {
//				return page, false
//			}
//			return page + 1, true
//		},
//	)
//	return api(1, dst)
//}
