package kugou

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/AynaLivePlayer/miaosic/utils"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
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
	cookie    map[string]string
	appid     string
	clientver string
	signkey   string
	dfid      string // dfid. default should be "-"
	client    *resty.Client
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
		client:    resty.New().SetTimeout(3 * time.Second),
	}
	if useLite {
		pvdr.appid = appidLite
		pvdr.clientver = clientverLite
		pvdr.signkey = signkeyLite
	}
	pvdr.cookie = make(map[string]string)
	return pvdr
}

func (k *Kugou) GetName() string {
	return "kugou"
}

func (k *Kugou) Search(keyword string, page, size int) ([]miaosic.MediaInfo, error) {
	resp, err := k.client.R().
		SetHeaders(header).
		SetQueryParams(map[string]string{
			"keyword":  keyword,
			"page":     fmt.Sprintf("%d", page),
			"pagesize": fmt.Sprintf("%d", size),
		}).
		Get("http://mobilecdn.kugou.com/api/v3/search/song")
	if err != nil {
		return nil, err
	}
	respResult := gjson.ParseBytes(resp.Body())
	if respResult.Get("errcode").Int() != 0 {
		return nil, errors.New("kugou: search api error" + respResult.Get("error").String())
	}
	//fmt.Println(resp.String())
	// Assuming data contains a list of search results
	result := make([]miaosic.MediaInfo, 0)
	for _, r := range respResult.Get("data.info").Array() {
		media := miaosic.MediaInfo{
			Title:  r.Get("songname").String(),
			Album:  r.Get("album_name").String(),
			Cover:  miaosic.Picture{},
			Artist: r.Get("singername").String(),
			Meta: miaosic.MetaData{
				Provider:   k.GetName(),
				Identifier: r.Get("hash").String(),
			},
		}
		result = append(result, media)
	}
	return result, nil
}

func (k *Kugou) GetMediaInfo(meta miaosic.MetaData) (miaosic.MediaInfo, error) {
	data := map[string]interface{}{
		"appid":            k.appid,
		"area_code":        1,
		"behavior":         "play",
		"clientver":        k.clientver,
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
	resp, err := k.client.R().
		SetHeaders(map[string]string{"Content-Type": "application/json", "x-router": "media.store.kugou.com"}).
		SetBody(dataRaw).
		Post("http://media.store.kugou.com/v2/get_res_privilege/lite")
	if err != nil {
		return miaosic.MediaInfo{}, err
	}
	result := gjson.ParseBytes(resp.Body())
	media := miaosic.MediaInfo{Meta: meta}
	if result.Get("data.0.name").String() == "" {
		return miaosic.MediaInfo{}, errors.New("failed to find required data")
	}
	artist := result.Get("data.0.singername").String()
	media.Title = strings.TrimPrefix(result.Get("data.0.name").String(), artist+" - ")
	media.Artists = strings.Split(artist, "、")
	media.Artist = strings.Join(media.Artists, ",")
	media.Album = result.Get("data.0.albumname").String()
	media.Cover.Url = strings.Replace(result.Get("data.0.info.image").String(), "{size}", result.Get("data.0.info.imgsize.0").String(), 1)
	return media, nil
}

var kugouIdRegex = regexp.MustCompile("^[0-9a-zA-Z]{32,32}$")

func (k *Kugou) MatchMedia(uri string) (miaosic.MetaData, bool) {
	if id := kugouIdRegex.FindString(uri); id != "" {
		return miaosic.MetaData{
			Provider:   k.GetName(),
			Identifier: strings.ToLower(id),
		}, true
	}
	return miaosic.MetaData{}, false
}

// todo using new api https://github.com/MakcRe/KuGouMusicApi/blob/main/module/song_url_new.js
func (k *Kugou) GetMediaUrl(meta miaosic.MetaData, quality miaosic.Quality) ([]miaosic.MediaUrl, error) {
	//albumId := jsonResp.Get("data.0.audio_id").String()
	currentUnix := time.Now().UnixMilli()
	qualityStr := string(k.MapQuality(quality))
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
		"ppage_id":       "463467626,350369493,788954147",
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
	urlResp, err := k.client.R().
		SetHeaders(map[string]string{
			"x-router":   "tracker.kugou.com",
			"dfid":       k.dfid,
			"mid":        getMD5Hash(k.dfid),
			"clienttime": fmt.Sprintf("%d", currentUnix),
		}).
		SetQueryParams(k.stringifyParams(data)).
		Get("https://gateway.kugou.com/v5/url")
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
		urls = append(urls, miaosic.NewMediaUrl(value.String(), miaosic.Quality(qualityStr)))
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
	resp, err := k.client.R().
		SetHeaders(header).
		SetQueryParams(k.stringifyParams(map[string]any{
			"hash": meta.Identifier,
		})).
		Get("http://lyrics.kugou.com/search?ver=1&man=yes&client=pc&keyword=&duration=")
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
	lyricResp, err := k.client.R().
		SetHeaders(header).
		SetQueryParams(k.stringifyParams(map[string]any{
			"id":        candidates[0].Get("id").String(),
			"accesskey": candidates[0].Get("accesskey").String(),
		})).
		Get("http://lyrics.kugou.com/download?ver=1&client=pc&fmt=lrc&charset=utf8")
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
	return []miaosic.Lyrics{utils.ParseLyricWithLangDetection(string(lyricdata))}, nil
}
