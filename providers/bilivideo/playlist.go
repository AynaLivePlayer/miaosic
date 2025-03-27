package bilivideo

import (
	"errors"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/aynakeya/deepcolor/dphttp"
	"github.com/tidwall/gjson"
	"regexp"
	"strconv"
	"strings"
)

const (
	playlistCollection = "coll"
	playlistFav        = "fav"
)

func fetchParsedResult[P dphttp.ParserResultType](requester dphttp.IRequester, request *dphttp.Request, parserFunc dphttp.ParserFunc[P]) (P, error) {
	httpResp, err := requester.HTTP(request)
	if err != nil {
		return *new(P), err
	}
	return parserFunc(httpResp)
}

var playlistCollectionRegex = regexp.MustCompile(`space.bilibili.com/(\d+)/channel/collectiondetail\?sid=(\d+)`)
var playlistCollection1Regex = regexp.MustCompile(`space\.bilibili\.com/(\d+)/lists/(\d+)`)
var playlistFavRegex = regexp.MustCompile(`space.bilibili.com/(\d+)/favlist\?fid=(\d+)`)

func makePlaylistId(ptype string, id string) string {
	return ptype + "_" + id
}

func parsePlaylistId(pid string) (string, string) {
	parts := strings.SplitN(pid, "_", 2)
	if parts[0] != playlistCollection && parts[0] != playlistFav {
		return "", ""
	}
	if _, err := strconv.ParseInt(parts[1], 10, 64); err != nil {
		return "", ""
	}
	return parts[0], parts[1]
}

func (n *BilibiliVideo) MatchPlaylist(uri string) (miaosic.MetaData, bool) {
	if playlistCollectionRegex.MatchString(uri) {
		matches := playlistCollectionRegex.FindStringSubmatch(uri)
		return miaosic.MetaData{
			Provider:   n.GetName(),
			Identifier: makePlaylistId(playlistCollection, matches[2]),
		}, true
	}
	if playlistCollection1Regex.MatchString(uri) {
		matches := playlistCollection1Regex.FindStringSubmatch(uri)
		return miaosic.MetaData{
			Provider:   n.GetName(),
			Identifier: makePlaylistId(playlistCollection, matches[2]),
		}, true
	}
	if playlistFavRegex.MatchString(uri) {
		matches := playlistFavRegex.FindStringSubmatch(uri)
		return miaosic.MetaData{
			Provider:   n.GetName(),
			Identifier: makePlaylistId(playlistFav, matches[2]),
		}, true
	}
	return miaosic.MetaData{}, false
}

var collApi = "https://api.bilibili.com/x/polymer/web-space/seasons_archives_list?mid=0&season_id=%s&sort_reverse=false&page_num=%d&page_size=30"

func (n *BilibiliVideo) getCollectionPlaylist(id string) (*miaosic.Playlist, error) {

	playlist := &miaosic.Playlist{
		Meta:   miaosic.MetaData{n.GetName(), makePlaylistId(playlistCollection, id)},
		Medias: make([]miaosic.MediaInfo, 0),
		Title:  "Bilibili Collection " + id,
	}
	for page := 1; page <= 50; page++ {
		uri := fmt.Sprintf(collApi, id, page)
		resp, err := miaosic.Requester.Get(uri, biliHeaders)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() != 200 {
			return nil, miaosic.ErrorExternalApi
		}
		result := gjson.ParseBytes(resp.Body())
		if result.Get("code").Int() != 0 {
			return nil, errors.New("bilivideo: " + result.Get("message").String())
		}
		archives := result.Get("data.archives")
		if len(archives.Array()) == 0 {
			playlist.Title = result.Get("data.meta.name").String()
			break
		}
		archives.ForEach(func(key, value gjson.Result) bool {
			playlist.Medias = append(playlist.Medias, miaosic.MediaInfo{
				Title:  value.Get("title").String(),
				Cover:  miaosic.Picture{Url: value.Get("pic").String()},
				Artist: id,
				Meta: miaosic.MetaData{
					Provider:   n.GetName(),
					Identifier: value.Get("bvid").String(),
				},
			})
			return true
		})
	}
	if len(playlist.Medias) == 0 {
		return nil, errors.New("bilivideo: no media found")
	}
	return playlist, nil
}

var favApi = "https://api.bilibili.com/x/v3/fav/resource/list?media_id=%s&pn=%d&ps=20&keyword=&order=mtime&type=0&tid=0&platform=web"

func (n *BilibiliVideo) getFavPlaylist(id string) (*miaosic.Playlist, error) {
	playlist := &miaosic.Playlist{
		Meta:   miaosic.MetaData{n.GetName(), makePlaylistId(playlistFav, id)},
		Medias: make([]miaosic.MediaInfo, 0),
		Title:  "Bilibili Fav " + id,
	}
	for page := 1; page < 51; page++ {
		uri := fmt.Sprintf(favApi, id, page)
		resp, err := miaosic.Requester.Get(uri, biliHeaders)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() != 200 {
			return nil, miaosic.ErrorExternalApi
		}
		result := gjson.ParseBytes(resp.Body())
		if result.Get("code").Int() != 0 {
			return nil, errors.New("bilivideo: " + result.Get("message").String())
		}
		medias := result.Get("data.medias")
		if len(medias.Array()) == 0 {
			playlist.Title = result.Get("data.info.title").String()
			break
		}
		medias.ForEach(func(key, value gjson.Result) bool {
			title := value.Get("title").String()
			if title == "已失效视频" || title == "" {
				return true
			}
			playlist.Medias = append(playlist.Medias, miaosic.MediaInfo{
				Title:  value.Get("title").String(),
				Cover:  miaosic.Picture{Url: value.Get("cover").String()},
				Artist: value.Get("upper.name").String(),
				Meta: miaosic.MetaData{
					Provider:   n.GetName(),
					Identifier: value.Get("bvid").String() + "?p=" + value.Get("page").String(),
				},
			})
			return true
		})
		if !result.Get("data.has_more").Bool() {
			playlist.Title = result.Get("data.info.title").String()
			break
		}
	}
	return playlist, nil
}

func (n *BilibiliVideo) GetPlaylist(meta miaosic.MetaData) (*miaosic.Playlist, error) {
	ptype, id := parsePlaylistId(meta.Identifier)
	if ptype == "" {
		return nil, errors.New("bilivideo: invalid playlist identifier")
	}
	if ptype == playlistCollection {
		return n.getCollectionPlaylist(id)
	}
	return n.getFavPlaylist(id)
}
