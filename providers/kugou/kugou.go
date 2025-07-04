package kugou

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/AynaLivePlayer/miaosic/providers"
	"github.com/aynakeya/deepcolor"
	"github.com/aynakeya/deepcolor/dphttp"
	"github.com/tidwall/gjson"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
)

var header = map[string]string{
	"user-agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
}

// basically kugou have two channels
// 1. lite
// 2. standard
// each channel have its own appid and clientver.
// and they also have different signkey.
// moreover, they even have different vip. which means lite vip can't access standard vip.

type Kugou struct {
	providers.DeepcolorProvider
	cookie    map[string]string
	appid     string
	clientver string
	signkey   string
	dfid      string // dfid. default should be "-"
}

func (k *Kugou) Qualities() []miaosic.Quality {
	return []miaosic.Quality{
		Quality128k,
		Quality320k,
		QualityFlac,
		QualityHigh,
		QualityViperTape,
		QualityViperClear,
		QualityViperHiFi,
		QualityViperAtmosphere,
	}
}

func (k *Kugou) cookieString() string {
	var cookieStr strings.Builder
	for key, value := range k.cookie {
		cookieStr.WriteString(fmt.Sprintf("%s=%s;", key, value))
	}
	return cookieStr.String()
}

func NewKugou(useLite bool) *Kugou {
	pvdr := &Kugou{
		appid:     appid,
		clientver: clientver,
		signkey:   signkey,
		dfid:      "-",
	}
	if useLite {
		pvdr.appid = appidLite
		pvdr.clientver = clientverLite
		pvdr.signkey = signkeyLite
	}
	pvdr.InfoApi = deepcolor.CreateApiResultFunc(
		func(meta miaosic.MetaData) (*dphttp.Request, error) {
			uri, _ := url.Parse("http://media.store.kugou.com/v2/get_res_privilege/lite")
			data := map[string]interface{}{
				"appid":            pvdr.appid,
				"area_code":        1,
				"behavior":         "play",
				"clientver":        pvdr.clientver,
				"need_hash_offset": 1,
				"relate":           1,
				"support_verify":   1,
				"resource": []map[string]interface{}{
					{
						"type":     "audio",
						"page_id":  0,
						"hash":     meta.Identifier,
						"album_id": 0,
					},
				},
				"qualities": []string{"128", "320", "flac", "high", "viper_atmos", "viper_tape"},
			}
			dataRaw, _ := json.Marshal(data)
			return &dphttp.Request{
				Method:  http.MethodPost,
				Url:     uri,
				Header:  map[string]string{"Content-Type": "application/json", "x-router": "media.store.kugou.com"},
				Data:    dataRaw,
				Timeout: 3,
			}, nil
		},
		deepcolor.ParserGJson,
		func(result *gjson.Result, media *miaosic.MediaInfo) error {
			if result.Get("data.0.name").String() == "" {
				return errors.New("failed to find required data")
			}
			media.Title = result.Get("data.0.name").String()
			media.Artist = result.Get("data.0.singername").String()
			media.Album = result.Get("data.0.albumname").String()
			media.Cover.Url = strings.Replace(result.Get("data.0.info.image").String(), "{size}", result.Get("data.0.info.imgsize.0").String(), 1)
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
	pvdr.cookie = make(map[string]string)
	return pvdr
}

func (k *Kugou) GetName() string {
	return "kugou"
}

var kugouIdRegex = regexp.MustCompile("^[0-9a-z]{32,32}$")

func (k *Kugou) MatchMedia(uri string) (miaosic.MetaData, bool) {
	if id := kugouIdRegex.FindString(uri); id != "" {
		return miaosic.MetaData{
			Provider:   k.GetName(),
			Identifier: id,
		}, true
	}
	return miaosic.MetaData{}, false
}

func (k *Kugou) quality2str(quality miaosic.Quality) string {
	if slices.Contains(k.Qualities(), quality) {
		return string(quality)
	}
	var qualityStr string
	if strings.HasPrefix(string(quality), "magic_") {
		qualityStr = string(quality)
	} else {
		switch quality {
		case miaosic.Quality128k:
			qualityStr = "128"
		case miaosic.Quality192k:
		case miaosic.Quality256k:
		case miaosic.Quality320k:
			qualityStr = "320"
		case miaosic.QualityHQ:
			qualityStr = "high"
		case miaosic.QualitySQ:
			qualityStr = "flac"
		default:
			qualityStr = "320"
		}
	}
	return qualityStr
}

func (k *Kugou) GetMediaUrl(meta miaosic.MetaData, quality miaosic.Quality) ([]miaosic.MediaUrl, error) {
	//albumId := jsonResp.Get("data.0.audio_id").String()
	currentUnix := time.Now().UnixMilli()
	qualityStr := k.quality2str(quality)
	data := map[string]any{
		"album_audio_id": 0,
		"appid":          k.appid,
		"clientver":      k.clientver,
		"clienttime":     fmt.Sprintf("%d", currentUnix),
		"area_code":      1,
		"hash":           meta.Identifier,
		"vipType":        0,
		"vipToken":       "",
		"behavior":       "play",
		"pid":            2,
		"cmd":            26,
		"pidversion":     3001,
		"isFreePart":     0,
		"album_id":       0,
		"ssa_flag":       "is_fromtrack",
		"version":        11709,
		"page_id":        312258376,
		"quality":        qualityStr,
		"ppgea_id":       "463467626,350369493,788954147",
		"cdnBackup":      1,
		"kcard":          0,
		"ptype":          0,
		"key":            signKey(k.appid, meta.Identifier, getMD5Hash(k.dfid), "0"),
		"dfid":           k.dfid,
		"mid":            getMD5Hash(k.dfid),
	}
	if token, ok := k.cookie["token"]; ok {
		data["token"] = token
	}
	if userId, ok := k.cookie["userid"]; ok {
		data["userid"] = userId
		data["key"] = signKey(k.appid, meta.Identifier, getMD5Hash(k.dfid), userId)
	}
	data["signature"] = signatureAndroidParams(k.signkey, data, "")
	urlReq, _ := deepcolor.NewGetRequestWithQuery(
		"https://gateway.kugou.com/v5/url",
		data, map[string]string{
			"x-router":   "tracker.kugou.com",
			"dfid":       k.dfid,
			"mid":        getMD5Hash(k.dfid),
			"clienttime": fmt.Sprintf("%d", currentUnix),
		},
	)
	urlResp, err := miaosic.Requester.HTTP(urlReq)
	if err != nil {
		return nil, err
	}
	//fmt.Println(urlResp.String())
	urlJson := gjson.ParseBytes(urlResp.Body())
	if !urlJson.Get("url").Exists() {
		return nil, errors.New("kugou: failed to get media url, might be vip only")
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
