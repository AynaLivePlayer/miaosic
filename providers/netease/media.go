package netease

import (
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/AynaLivePlayer/miaosic/utils"
	neteaseApi "github.com/XiaoMengXinX/Music163Api-Go/api"
	neteaseTypes "github.com/XiaoMengXinX/Music163Api-Go/types"
	neteaseUtil "github.com/XiaoMengXinX/Music163Api-Go/utils"
	"github.com/spf13/cast"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type Netease struct {
	ReqData        neteaseUtil.RequestData
	IdRegex0       *regexp.Regexp
	IdRegex1       *regexp.Regexp
	PlaylistRegex0 *regexp.Regexp
	PlaylistRegex1 *regexp.Regexp
	deviceId       string
}

func (n *Netease) Qualities() []miaosic.Quality {
	return []miaosic.Quality{
		QualityStandard,
		QualityHigher,
		QualityExHigh,
		QualityLossless,
		QualityHiRes,
		QualityJyMaster,
	}
}

func NewNetease() *Netease {
	deviceId, err := generateRandomString(32)
	if err != nil {
		deviceId = "00000000000000000000000000000000"
	}
	return &Netease{
		ReqData: neteaseUtil.RequestData{
			Headers: neteaseUtil.Headers{
				{
					"X-Real-IP",
					"118.88.88.88",
				},
			},
			Cookies: []*http.Cookie{
				{
					Name:  "deviceId",
					Value: deviceId,
				},
			},
		},
		IdRegex0:       regexp.MustCompile("^[0-9]+$"),
		IdRegex1:       regexp.MustCompile("^wy[0-9]+$"),
		PlaylistRegex0: regexp.MustCompile("^[0-9]+$"),
		// https://music.163.com/playlist?id=2382819181&userid=95906480
		PlaylistRegex1: regexp.MustCompile("playlist\\?id=[0-9]+"),
		deviceId:       deviceId,
	}
}

func (n *Netease) GetName() string {
	return "netease"
}

func (n *Netease) MatchMedia(uri string) (miaosic.MetaData, bool) {
	if id := n.IdRegex0.FindString(uri); id != "" {
		return miaosic.MetaData{
			Provider:   n.GetName(),
			Identifier: id,
		}, true
	}
	if id := n.IdRegex1.FindString(uri); id != "" {
		return miaosic.MetaData{
			Provider:   n.GetName(),
			Identifier: id[2:],
		}, true
	}
	return miaosic.MetaData{}, false
}

func (n *Netease) Search(keyword string, page, size int) ([]miaosic.MediaInfo, error) {
	rawResult, err := neteaseApi.SearchSong(
		n.ReqData,
		neteaseApi.SearchSongConfig{
			Keyword: keyword,
			Limit:   size,
			Offset:  (page - 1) * size,
		})
	if err != nil || rawResult.Code != 200 {
		return nil, miaosic.ErrorExternalApi
	}
	medias := make([]miaosic.MediaInfo, 0)
	for _, song := range rawResult.Result.Songs {
		artists := make([]string, 0)
		for _, a := range song.Artists {
			artists = append(artists, a.Name)
		}
		medias = append(medias, miaosic.MediaInfo{
			Title:  song.Name,
			Artist: strings.Join(artists, ","),
			Cover:  miaosic.Picture{},
			Album:  song.Album.Name,
			Meta: miaosic.MetaData{
				Provider:   n.GetName(),
				Identifier: strconv.Itoa(song.Id),
			},
		})
	}
	return medias, nil
}

func _neteaseGetArtistNames(data neteaseTypes.SongDetailData) string {
	artists := make([]string, 0)
	for _, a := range data.Ar {
		artists = append(artists, a.Name)
	}
	return strings.Join(artists, ",")
}

func (n *Netease) GetMediaInfo(meta miaosic.MetaData) (media miaosic.MediaInfo, err error) {
	result, err := neteaseApi.GetSongDetail(
		n.ReqData,
		[]int{cast.ToInt(meta.Identifier)})
	if err != nil || result.Code != 200 {
		return miaosic.MediaInfo{}, miaosic.ErrorExternalApi
	}
	if len(result.Songs) == 0 {
		return miaosic.MediaInfo{}, miaosic.ErrorExternalApi
	}
	media.Title = result.Songs[0].Name
	media.Cover.Url = result.Songs[0].Al.PicUrl
	media.Album = result.Songs[0].Al.Name
	media.Artist = _neteaseGetArtistNames(result.Songs[0])
	media.Meta = meta
	return media, nil
}

func (n *Netease) quality2str(quality miaosic.Quality) string {
	if slices.Contains(n.Qualities(), quality) {
		return string(quality)
	}
	return "standard"
}

func (n *Netease) GetMediaUrl(meta miaosic.MetaData, quality miaosic.Quality) ([]miaosic.MediaUrl, error) {
	result, err := neteaseApi.GetSongURL(
		n.ReqData,
		neteaseApi.SongURLConfig{
			Ids:   []int{cast.ToInt(meta.Identifier)},
			Level: n.quality2str(quality),
		})
	if err != nil || result.Code != 200 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("miaosic (netease): error errcode %d", result.Code)
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("miaosic (netease): no result found")
	}
	if result.Data[0].Code != 200 {
		return nil, fmt.Errorf("miaosic (netease): netease code not equal to 200, might no copyright")
	}
	urls := make([]miaosic.MediaUrl, 0)
	for _, u := range result.Data {
		qualityStr, ok := u.Level.(string)
		if !ok {
			qualityStr = "unknown"
		}
		urls = append(urls, miaosic.MediaUrl{
			Url:     u.Url,
			Quality: miaosic.Quality(qualityStr),
		})
	}
	return urls, nil
}

func (n *Netease) GetMediaLyric(meta miaosic.MetaData) ([]miaosic.Lyrics, error) {
	result, err := neteaseApi.GetSongLyric(n.ReqData, cast.ToInt(meta.Identifier))
	if err != nil || result.Code != 200 {
		return nil, miaosic.ErrorExternalApi
	}
	lrcs := make([]miaosic.Lyrics, 0)
	if result.Lrc.Lyric != "" {
		lrcs = append(lrcs, utils.ParseLyricWithLangDetection(result.Lrc.Lyric))
	}
	if result.Tlyric.Lyric != "" {
		lrcs = append(lrcs, utils.ParseLyricWithLangDetection(result.Tlyric.Lyric))
	}
	if result.Klyric.Lyric != "" {
		lrcs = append(lrcs, miaosic.ParseLyrics("karaoke", result.Klyric.Lyric))
	}
	return lrcs, nil
}
