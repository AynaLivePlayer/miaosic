package qq

import (
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/AynaLivePlayer/miaosic/utils"
	"github.com/tidwall/gjson"
	"regexp"
	"slices"
	"strings"
)

type ApiConfig struct {
	Version     string
	VersionCode int
	EnableSign  bool
	Endpoint    string
	EncEndpoint string
}

type QQMusicProvider struct {
	cfg            ApiConfig
	device         *Device
	cred           *Credential
	header         map[string]string
	qimeiUpdated   bool //i don't care concurrence
	tokenRefreshed bool
	channel        string // "qq" or "wechat"
}

func (p *QQMusicProvider) GetName() string {
	return "qq"
}

func (p *QQMusicProvider) Qualities() []miaosic.Quality {
	return []miaosic.Quality{
		QualityMaster, QualityAtmos2, QualityAtmos51,
		QualityFLAC,
		QualityOGG640, QualityOGG320, QualityOGG192, QualityOGG96,
		QualityMP3320, QualityMP3128, QualityACC192,
		QualityACC96, QualityACC48,
	}
}

func NewQQMusicProvider(channel string) *QQMusicProvider {
	if channel != "qq" && channel != "wechat" {
		channel = "qq"
	}
	val := &QQMusicProvider{
		cfg: ApiConfig{
			Version:     "13.2.5.8",
			VersionCode: 13020508,
			EnableSign:  true,
			Endpoint:    "https://u.y.qq.com/cgi-bin/musics.fcg",
			EncEndpoint: "https://u.y.qq.com/cgi-bin/musics.fcg",
		},
		cred:   NewCredential(),
		device: NewDevice(),
		header: map[string]string{
			"host":       "y.qq.com",
			"user-agent": "Mozilla/5.0 (Windows NT 11.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36 Edg/116.0.1938.54",
		},
		qimeiUpdated:   false,
		tokenRefreshed: false,
		channel:        channel,
	}
	return val
}

var idRegexp = regexp.MustCompile(`^\d{3,3}[0-9A-Za-z]{11,11}$`)

func (p *QQMusicProvider) MatchMedia(uri string) (miaosic.MetaData, bool) {
	if id := idRegexp.FindString(uri); id != "" {
		return miaosic.MetaData{
			Provider:   p.GetName(),
			Identifier: id,
		}, true
	}
	return miaosic.MetaData{}, false
}

func (p *QQMusicProvider) Search(keyword string, page, size int) ([]miaosic.MediaInfo, error) {
	params := map[string]interface{}{
		"searchid":     getSearchID(),
		"query":        keyword,
		"search_type":  "item_song",
		"num_per_page": size,
		"page_num":     page,
		"highlight":    1, // 1: true
		"grp":          1, // 1: true
	}
	data, err := p.makeApiRequest("music.search.SearchCgiService", "DoSearchForQQMusicMobile", params)
	if err != nil {
		return nil, err
	}

	var medias []miaosic.MediaInfo
	data.Get("data.body.item_song").ForEach(func(key, info gjson.Result) bool {
		title := info.Get("title").String()
		mid := info.Get("mid").String()

		var artistNames []string
		info.Get("singer").ForEach(func(key, value gjson.Result) bool {
			name := value.Get("name").String()
			if name != "" {
				artistNames = append(artistNames, name)
			}
			return true
		})
		artist := strings.Join(artistNames, ",")
		albumMid := info.Get("album.mid").String()
		coverURL := ""
		albumSize := 500 //[150, 300, 500, 800]:
		if albumMid != "" {
			coverURL = fmt.Sprintf("https://y.qq.com/music/photo_new/T002R%dx%dM000%s.jpg", albumSize, albumSize, albumMid)
		}

		medias = append(medias, miaosic.MediaInfo{
			Title:  title,
			Artist: artist,
			Album:  info.Get("album.title").String(),
			Cover:  miaosic.Picture{Url: coverURL},
			Meta:   miaosic.MetaData{Provider: p.GetName(), Identifier: mid},
		})
		return true
	})
	return medias, nil
}

func (p *QQMusicProvider) GetMediaInfo(meta miaosic.MetaData) (miaosic.MediaInfo, error) {
	params := map[string]interface{}{
		"mids":         []string{meta.Identifier},
		"types":        []int{0},
		"modify_stamp": []int{0},
		"ctx":          0,
		"client":       1,
	}

	data, err := p.makeApiRequest("music.trackInfo.UniformRuleCtrl", "CgiGetTrackInfo", params)
	if err != nil {
		return miaosic.MediaInfo{}, err
	}

	track := data.Get("data.tracks.0")
	if !track.Exists() {
		return miaosic.MediaInfo{}, fmt.Errorf("miaosic (qq): song not found")
	}
	title := track.Get("title").String()
	mid := track.Get("mid").String()
	albumMid := track.Get("album.mid").String()
	albumTitle := track.Get("album.title").String()

	var artistNames []string
	track.Get("singer").ForEach(func(_, singer gjson.Result) bool {
		name := singer.Get("name").String()
		if name != "" {
			artistNames = append(artistNames, name)
		}
		return true
	})
	artist := strings.Join(artistNames, ",")

	coverURL := ""
	if albumMid != "" {
		coverURL = fmt.Sprintf("https://y.qq.com/music/photo_new/T002R500x500M000%s.jpg", albumMid)
	}

	return miaosic.MediaInfo{
		Title:  title,
		Artist: artist,
		Album:  albumTitle,
		Cover:  miaosic.Picture{Url: coverURL},
		Meta:   miaosic.MetaData{Provider: p.GetName(), Identifier: mid},
	}, nil
}

func (p *QQMusicProvider) asQQQuality(quality miaosic.Quality) miaosic.Quality {
	if slices.Contains(p.Qualities(), quality) {
		return quality
	}
	return QualityMP3320
}

func (p *QQMusicProvider) GetMediaUrl(meta miaosic.MetaData, quality miaosic.Quality) ([]miaosic.MediaUrl, error) {
	var module, method string
	quality = p.asQQQuality(quality)
	if isEncryptedQuality(quality) {
		module = "music.vkey.GetEVkey"
		method = "CgiGetEVkey"
	} else {
		module = "music.vkey.GetVkey"
		method = "UrlGetVkey"
	}

	qs := strings.Split(string(quality), ".")
	domain := "https://isure.stream.qqmusic.qq.com/"

	params := map[string]interface{}{
		"filename": []string{fmt.Sprintf("%sOvO%sQwQ.%s", qs[0], meta.Identifier, qs[1])},
		"guid":     getGuid(),
		"songmid":  []string{meta.Identifier},
		"songtype": []int{0},
	}

	data, err := p.makeApiRequest(module, method, params)
	if err != nil {
		return nil, err
	}
	wifiurl := data.Get("data.midurlinfo.0.wifiurl").String()
	if wifiurl == "" {
		return nil, fmt.Errorf("miaosic (qq): wifiurl not found, might require vip/no copyright")
	}
	result := []miaosic.MediaUrl{
		miaosic.MediaUrl{
			Url:     domain + wifiurl,
			Quality: quality,
		},
	}
	return result, err
}

func (p *QQMusicProvider) GetMediaLyric(meta miaosic.MetaData) ([]miaosic.Lyrics, error) {
	resp, err := p.makeApiRequest("music.musichallSong.PlayLyricInfo", "GetPlayLyricInfo", map[string]any{
		"songMid": meta.Identifier,
		"crypt":   1,
		"ct":      11,
		"cv":      13020508,
		"lrc_t":   0,
		"qrc":     0,
		"qrc_t":   0,
		"roma":    1,
		"roma_t":  0,
		"trans":   1,
		"trans_t": 0,
		"type":    1,
	})
	if err != nil {
		return nil, err
	}

	result := make([]miaosic.Lyrics, 0)

	if lyricEnc := resp.Get("data.lyric").String(); lyricEnc != "" {
		lyric, err := qrcDecrypt(lyricEnc)
		if err == nil {
			result = append(result, utils.ParseLyricWithLangDetection(lyric))
		} else {
			fmt.Println(err)
		}
	}
	if lyricEnc := resp.Get("data.trans").String(); lyricEnc != "" {
		lyric, err := qrcDecrypt(lyricEnc)
		if err == nil {
			result = append(result, utils.ParseLyricWithLangDetection(lyric))
		}
	}
	if lyricEnc := resp.Get("data.roma").String(); lyricEnc != "" {
		lyric, err := qrcDecrypt(lyricEnc)
		if err == nil {
			result = append(result, utils.ParseLyricWithLangDetection(lyric))
		}
	}
	return result, nil
}
