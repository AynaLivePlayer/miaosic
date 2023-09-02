package kuwo

import (
	"fmt"
	"github.com/aynakeya/deepcolor"
	"github.com/aynakeya/deepcolor/dphttp"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"html"
	"math"
	"math/rand"
	"miaosic"
	"miaosic/providers"
	"regexp"
	"strconv"
	"strings"
)

type Kuwo struct {
	providers.DeepcolorProvider
	requester      dphttp.IRequester
	PlaylistRegex0 *regexp.Regexp
	PlaylistRegex1 *regexp.Regexp
	IdRegex0       *regexp.Regexp
	IdRegex1       *regexp.Regexp
	header         map[string]string
}

func NewKuwo(requester dphttp.IRequester) *Kuwo {
	kw := &Kuwo{
		requester:      requester,
		PlaylistRegex0: regexp.MustCompile("[0-9]+"),
		PlaylistRegex1: regexp.MustCompile("playlist/[0-9]+"),
		IdRegex0:       regexp.MustCompile("^[0-9]+"),
		IdRegex1:       regexp.MustCompile("^kw[0-9]+"),
	}
	kw.initToken()
	kw.InfoFunc = kw.buildInfoApi()
	kw.FileFunc = kw.buildFileApi()
	kw.LyricFunc = kw.buildLyricApi()
	kw.PlaylistFunc = kw.playlistApi
	kw.SearchFunc = kw.buildSearchApi()
	return kw
}

func (k *Kuwo) GetName() string {
	return "kuwo"
}

func (k *Kuwo) MatchMedia(keyword string) *miaosic.Media {
	if id := k.IdRegex0.FindString(keyword); id != "" {
		return &miaosic.Media{
			Meta: miaosic.MediaMeta{
				Provider:   k.GetName(),
				Identifier: id,
			},
		}
	}
	if id := k.IdRegex1.FindString(keyword); id != "" {
		return &miaosic.Media{
			Meta: miaosic.MediaMeta{
				Provider:   k.GetName(),
				Identifier: id[2:],
			},
		}
	}
	return nil
}

func (k *Kuwo) MatchPlaylist(uri string) *miaosic.Playlist {
	var id string
	id = k.PlaylistRegex0.FindString(uri)
	if id != "" {
		return &miaosic.Playlist{
			Meta: miaosic.MediaMeta{k.GetName(), id},
		}
	}
	id = k.PlaylistRegex1.FindString(uri)
	if id != "" {
		return &miaosic.Playlist{
			Meta: miaosic.MediaMeta{k.GetName(), id[9:]},
		}
	}
	return nil
}

func (k *Kuwo) generateSecret(t, e string) string {
	if e == "" {
		return ""
	}

	var n string
	for i := 0; i < len(e); i++ {
		n += strconv.Itoa(int(e[i]))
	}
	r := len(n) / 5
	o_0 := string(n[r]) + string(n[2*r]) + string(n[3*r]) + string(n[4*r])
	if 5*r < len(n) {
		o_0 += string(n[5*r])
	}
	o, _ := strconv.Atoi(o_0)
	l := int(math.Ceil(float64(len(e)) / 2.0))
	c := int(math.Pow(2, 31)) - 1

	if o < 2 {
		return ""
	}

	d := rand.Intn(100000000)
	//d := 80378195
	n += strconv.Itoa(d)

	var num1, num2 int64
	for len(n) > 10 {
		// stupid javascript
		if len(n[10:]) > 19 {
			num1 = 0
			num2 = cast.ToInt64(n[19 : 19+8])
		} else {
			num1 = cast.ToInt64(n[:10])
			num2 = cast.ToInt64(n[10:])
		}
		n = cast.ToString(num1 + num2)
	}

	nValue, _ := strconv.Atoi(n)
	nValue = (o*nValue + l) % c

	var h int
	var f string

	for i := 0; i < len(t); i++ {
		h = int(t[i]) ^ int(math.Floor(float64(nValue)/float64(c)*255))
		hexValue := strconv.FormatInt(int64(h), 16)

		if h < 16 {
			f += "0" + hexValue
		} else {
			f += hexValue
		}

		nValue = (o*nValue + l) % c
	}

	// d to hex string, if d length < 8, add 0 to head
	dHex := fmt.Sprintf("%x", d)
	if len(dHex) < 8 {
		dHex = strings.Repeat("0", 8-len(dHex)) + dHex
	}

	return f + dHex
}

func (k *Kuwo) initToken() {
	k.header = map[string]string{
		"cookie":  "Hm_Iuvt_cdb524f42f0cer9b268e4v7y734w5esq24=TN7FsbxFGt8y2sTb4tGnzhpD7StNfiRM",
		"secret":  k.generateSecret("TN7FsbxFGt8y2sTb4tGnzhpD7StNfiRM", "Hm_Iuvt_cdb524f42f0cer9b268e4v7y734w5esq24"),
		"referer": "http://www.kuwo.cn/",
	}
	//searchCookie, err := k.requester.Get("http://kuwo.cn/search/list?key=any", nil)
	//fmt.Println(searchCookie.Header(), err)
}

func (k *Kuwo) buildInfoApi() dphttp.ApiFunc[*miaosic.Media, *miaosic.Media] {
	return deepcolor.CreateApiFunc(
		k.requester,
		func(media *miaosic.Media) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithQuery(
				"http://www.kuwo.cn/api/www/music/musicInfo",
				[]string{"mid"}, k.header)([]string{media.Meta.Identifier})
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, media *miaosic.Media) error {
			if resp.Get("data.musicrid").String() == "" {
				return miaosic.ErrorExternalApi
			}
			media.Title = html.UnescapeString(resp.Get("data.name").String())
			media.Cover.Url = resp.Get("data.pic").String()
			media.Artist = resp.Get("data.artist").String()
			media.Album = resp.Get("data.album").String()
			return nil
		})
}

func (k *Kuwo) buildLyricApi() dphttp.ApiFunc[*miaosic.Media, *miaosic.Media] {
	return deepcolor.CreateApiFunc(
		k.requester,
		func(media *miaosic.Media) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithQuery(
				"http://m.kuwo.cn/newh5/singles/songinfoandlrc",
				[]string{"musicId"}, k.header)([]string{media.Meta.Identifier})
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, media *miaosic.Media) error {
			lrcs := make([]string, 0)
			resp.Get("data.lrclist").ForEach(func(key, value gjson.Result) bool {
				lrcs = append(lrcs, fmt.Sprintf("[00:%s]%s", value.Get("time").String(), value.Get("lineLyric").String()))

				return true
			})
			media.Lyric = strings.Join(lrcs, "\n")
			return nil
		})
}

func (k *Kuwo) buildSearchApi() dphttp.ApiFuncResult[string, []*miaosic.Media] {
	return deepcolor.CreateApiResultFunc(
		k.requester,
		func(keyword string) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithQuery(
				"http://www.kuwo.cn/api/www/search/searchMusicBykeyWord",
				[]string{"key", "pn", "rn"}, k.header)([]string{keyword, "1", "64"})
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, result *[]*miaosic.Media) error {
			resp.Get("data.list").ForEach(func(key, value gjson.Result) bool {
				*result = append(*result, &miaosic.Media{
					Title:  html.UnescapeString(value.Get("name").String()),
					Cover:  miaosic.Picture{Url: value.Get("pic").String()},
					Artist: value.Get("artist").String(),
					Album:  value.Get("album").String(),
					Meta: miaosic.MediaMeta{
						Provider:   k.GetName(),
						Identifier: value.Get("rid").String(),
					},
				})
				return true
			})
			return nil
		})
}

func (k *Kuwo) buildFileApi() dphttp.ApiFunc[*miaosic.Media, *miaosic.Media] {
	return deepcolor.CreateApiFunc(
		k.requester,
		func(media *miaosic.Media) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithSingleQuery(
				"http://antiserver.kuwo.cn/anti.s?type=convert_url&format=mp3&response=url",
				"rid", k.header)("MUSIC_" + media.Meta.Identifier)
		},
		deepcolor.ParserText,
		func(resp string, media *miaosic.Media) error {
			media.Url = resp
			return nil
		})
}

func (k *Kuwo) playlistApi(src *miaosic.Playlist, dst *miaosic.Playlist) error {
	dst.Medias = make([]*miaosic.Media, 0)
	api := deepcolor.CreateChainApiFunc(
		k.requester,
		func(page int) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithQuery(
				"http://www.kuwo.cn/api/www/playlist/playListInfo",
				[]string{"pid", "pn", "rn"}, k.header)([]string{src.Meta.Identifier, cast.ToString(page), "100"})
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, playlist *miaosic.Playlist) error {
			resp.Get("data.musicList").ForEach(func(key, value gjson.Result) bool {
				playlist.Medias = append(
					playlist.Medias,
					&miaosic.Media{
						Title:  html.UnescapeString(value.Get("name").String()),
						Artist: value.Get("artist").String(),
						Cover:  miaosic.Picture{Url: value.Get("pic").String()},
						Album:  value.Get("album").String(),
						Meta: miaosic.MediaMeta{
							Provider:   k.GetName(),
							Identifier: value.Get("rid").String(),
						},
					})
				return true
			})
			return nil
		},
		func(page int, resp *gjson.Result, playlist *miaosic.Playlist) (int, bool) {
			if resp.Get("code").String() != "200" {
				return page, false
			}
			cnt := int(resp.Get("data.total").Int())
			if cnt <= page*100 {
				return page, false
			}
			return page + 1, true
		},
	)
	return api(1, dst)
}
