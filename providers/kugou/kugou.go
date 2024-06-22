package kugou

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/AynaLivePlayer/miaosic/providers"
	"github.com/aynakeya/deepcolor"
	"github.com/aynakeya/deepcolor/dphttp"
	"github.com/tidwall/gjson"
	"strings"
)

var header = map[string]string{
	"user-agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
}

type Kugou struct {
	providers.DeepcolorProvider
}

func NewKugou() miaosic.MediaProvider {
	pvdr := &Kugou{}
	pvdr.InfoApi = deepcolor.CreateApiResultFunc(
		func(meta miaosic.MetaData) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithQuery(
				"https://m.kugou.com/app/i/getSongInfo.php?cmd=playInfo&hash=d9ca820cda8c4b88dea1d791d133d686",
				map[string]any{
					"hash": meta.Identifier,
				}, header,
			)
		},
		deepcolor.ParserGJson,
		func(result *gjson.Result, media *miaosic.MediaInfo) error {
			if result.Get("songName").String() == "" {
				return errors.New("failed to find required data")
			}
			media.Title = result.Get("songName").String()
			media.Artist = result.Get("singerName").String()
			media.Cover.Url = strings.Replace(result.Get("imgUrl").String(), "{size}", "500", 1)
			return nil
		})
	pvdr.SearchApi = deepcolor.CreateApiResultFunc(
		func(param providers.MediaSearchParam) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithQuery(
				"http://mobilecdn.kugou.com/api/v3/search/song?keyword=reol&page=1&pagesize=10",
				map[string]any{
					"keyword":  param.Keyword,
					"page":     param.Page,
					"pagesize": param.PageSize,
				},
				header)
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, result *[]miaosic.MediaInfo) error {
			if resp.Get("errcode").Int() != 0 {
				return errors.New("kugou: search api error" + resp.Get("error").String())
			}
			// Assuming data contains a list of search results
			for _, r := range resp.Get("data.info").Array() {
				media := miaosic.MediaInfo{
					Title:  r.Get("songname").String(),
					Cover:  miaosic.Picture{},
					Artist: r.Get("singername").String(),
					Meta: miaosic.MetaData{
						Provider:   pvdr.GetName(),
						Identifier: r.Get("hash").String(),
					},
				}
				*result = append(*result, media)
			}
			return nil
		})
	return pvdr
}

func (k *Kugou) GetName() string {
	return "kugou"
}

func (k *Kugou) MatchMedia(uri string) (miaosic.MetaData, bool) {
	return miaosic.MetaData{}, false
}

func (k *Kugou) GetMediaUrl(meta miaosic.MetaData, quality miaosic.Quality) ([]miaosic.MediaUrl, error) {
	hashesReq, _ := deepcolor.NewGetRequestWithQuery(
		"https://m.kugou.com/app/i/getSongInfo.php?cmd=playInfo&hash=d9ca820cda8c4b88dea1d791d133d686",
		map[string]any{
			"hash": meta.Identifier,
		}, header,
	)
	resp, err := miaosic.Requester.HTTP(hashesReq)
	if err != nil {
		return nil, err
	}
	jsonResp := gjson.ParseBytes(resp.Body())
	songHash := ""
	switch quality {
	case miaosic.Quality128k:
		songHash = jsonResp.Get("extra.128hash").String()
	case miaosic.Quality192k:
	case miaosic.Quality256k:
	case miaosic.Quality320k:
		songHash = jsonResp.Get("extra.320hash").String()
	case miaosic.QualityHQ:
		songHash = jsonResp.Get("extra.highhash").String()
	case miaosic.QualitySQ:
		songHash = jsonResp.Get("extra.sqhash").String()
	default:
		songHash = jsonResp.Get("extra.320hash").String()
		quality = miaosic.Quality320k
	}
	// fallback to default hash
	if songHash == "" {
		songHash = meta.Identifier
		quality = miaosic.QualityUnk
	}
	albumId := jsonResp.Get("album_id").String()
	urlReq, _ := deepcolor.NewGetRequestWithQuery(
		"http://trackercdn.kugou.com/i/v2/?appid=1005&pid=2&cmd=25&behavior=play",
		map[string]any{
			"hash":     songHash,
			"album_id": albumId,
			"key":      getMD5Hash(songHash + "kgcloudv2"),
		}, header,
	)
	urlResp, err := miaosic.Requester.HTTP(urlReq)
	if err != nil {
		return nil, err
	}
	urlJson := gjson.ParseBytes(urlResp.Body())
	if urlJson.Get("status").Int() != 1 {
		return nil, errors.New("kugou: failed to get media url")
	}
	urls := make([]miaosic.MediaUrl, 0)
	urlJson.Get("url").ForEach(func(key, value gjson.Result) bool {
		urls = append(urls, miaosic.NewMediaUrl(value.String(), quality))
		return true
	})
	return urls, nil
}

func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (k *Kugou) GetMediaLyric(meta miaosic.MetaData) ([]miaosic.Lyrics, error) {
	// http://lyrics.kugou.com/search?ver=1&man=yes&client=pc&keyword=&duration=&hash=c79c062ff4b362ac253031c6e577e722
	lyricReq, _ := deepcolor.NewGetRequestWithQuery(
		"http://lyrics.kugou.com/search?ver=1&man=yes&client=pc&keyword=&duration=",
		map[string]any{
			"hash": meta.Identifier,
		}, header)
	resp, err := miaosic.Requester.HTTP(lyricReq)
	if err != nil {
		return nil, err
	}
	jsonResp := gjson.ParseBytes(resp.Body())
	if jsonResp.Get("errcode").Int() != 200 {
		return nil, errors.New("kugou: failed to get lyric" + jsonResp.Get("errmsg").String())
	}
	candidates := jsonResp.Get("candidates").Array()
	if len(candidates) == 0 {
		return nil, errors.New("kugou: no lyric found")
	}
	lyricContentReq, _ := deepcolor.NewGetRequestWithQuery(
		"http://lyrics.kugou.com/download?ver=1&client=pc&id=&accesskey=&fmt=lrc&charset=utf8",
		map[string]any{
			"id":        candidates[0].Get("id").String(),
			"accesskey": candidates[0].Get("accesskey").String(),
		}, header)
	lyricResp, err := miaosic.Requester.HTTP(lyricContentReq)
	if err != nil {
		return nil, err
	}
	lyricJson := gjson.ParseBytes(lyricResp.Body())
	if lyricJson.Get("status").Int() != 200 {
		return nil, errors.New("kugou: failed to get lyric content")
	}
	lyricContent := lyricJson.Get("content").String()
	lyricdata, err := base64.StdEncoding.DecodeString(lyricContent)
	if err != nil {
		return nil, err
	}
	return []miaosic.Lyrics{miaosic.ParseLyrics("default", string(lyricdata))}, nil
}
