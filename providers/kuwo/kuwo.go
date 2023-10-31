package kuwo

import (
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/AynaLivePlayer/miaosic/providers"
	"github.com/aynakeya/deepcolor"
	"github.com/aynakeya/deepcolor/dphttp"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"html"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

type Kuwo struct {
	providers.DeepcolorProvider
	PlaylistRegex0 *regexp.Regexp
	PlaylistRegex1 *regexp.Regexp
	IdRegex0       *regexp.Regexp
	IdRegex1       *regexp.Regexp
	header         map[string]string
}

func NewKuwo() *Kuwo {
	kw := &Kuwo{
		PlaylistRegex0: regexp.MustCompile("[0-9]+"),
		PlaylistRegex1: regexp.MustCompile("playlist/[0-9]+"),
		IdRegex0:       regexp.MustCompile("^[0-9]+"),
		IdRegex1:       regexp.MustCompile("^kw[0-9]+"),
	}
	kw.initToken()
	kw.InfoApi = deepcolor.CreateApiResultFunc(
		func(meta miaosic.MediaMeta) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithSingleQuery(
				"http://www.kuwo.cn/api/www/music/musicInfo?httpsStatus=1",
				"mid", meta.Identifier, kw.header)
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, media *miaosic.MediaInfo) error {
			if resp.Get("data.musicrid").String() == "" {
				return miaosic.ErrorExternalApi
			}
			media.Title = html.UnescapeString(resp.Get("data.name").String())
			media.Cover.Url = resp.Get("data.pic").String()
			media.Artist = resp.Get("data.artist").String()
			media.Album = resp.Get("data.album").String()
			return nil
		})
	kw.FileApi = deepcolor.CreateApiResultFunc(
		func(param providers.FileApiParam) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithSingleQuery(
				"http://antiserver.kuwo.cn/anti.s?type=convert_url&format=mp3&response=url",
				"rid", "MUSIC_"+param.Meta.Identifier, kw.header)
		},
		deepcolor.ParserText,
		func(resp string, urls *[]miaosic.MediaUrl) error {
			*urls = []miaosic.MediaUrl{miaosic.NewMediaUrl(resp, miaosic.QualityUnk)}
			return nil
		})
	kw.LyricApi = deepcolor.CreateApiResultFunc(
		func(meta miaosic.MediaMeta) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithSingleQuery(
				"http://m.kuwo.cn/newh5/singles/songinfoandlrc",
				"musicId", meta.Identifier, kw.header)
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, lyrics *[]miaosic.Lyrics) error {
			lrcs := make([]string, 0)
			resp.Get("data.lrclist").ForEach(func(key, value gjson.Result) bool {
				lrcs = append(lrcs, fmt.Sprintf("[00:%s]%s", value.Get("time").String(), value.Get("lineLyric").String()))
				return true
			})
			if len(lrcs) == 0 {
				return miaosic.ErrorExternalApi
			}
			*lyrics = []miaosic.Lyrics{miaosic.ParseLyrics("default", strings.Join(lrcs, "\n"))}
			return nil
		})
	kw.SearchApi = deepcolor.CreateApiResultFunc(
		func(param providers.MediaSearchParam) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithQuery(
				"http://www.kuwo.cn/search/searchMusicBykeyWord/searchMusicBykeyWord?vipver=1&client=kt&ft=music&cluster=0&strategy=2012&encoding=utf8&rformat=json&mobi=1&issubtitle=1&show_copyright_off=1",
				map[string]any{
					"all": param.Keyword,
					"pn":  param.Page - 1,
					"rn":  param.PageSize,
				}, kw.header)
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, result *[]miaosic.MediaInfo) error {
			resp.Get("abslist").ForEach(func(key, value gjson.Result) bool {
				*result = append(*result, miaosic.MediaInfo{
					Title:  html.UnescapeString(value.Get("SONGNAME").String()),
					Cover:  miaosic.Picture{Url: "https://img2.kuwo.cn/star/albumcover/" + value.Get("web_albumpic_short").String()},
					Artist: value.Get("ARTIST").String(),
					Album:  value.Get("ALBUM").String(),
					Meta: miaosic.MediaMeta{
						Provider:   kw.GetName(),
						Identifier: value.Get("DC_TARGETID").String(),
					},
				})
				return true
			})
			return nil
		})
	//kw.PlaylistFunc = kw.playlistApi
	return kw
}

func (k *Kuwo) initToken() {
	k.header = map[string]string{
		"accept": "application/json, text/plain, */*",
		"cookie": "Hm_Iuvt_cdb524f42f0cer9b268e4v7y735ewrq2324=Ta28hMQmpTA2FCyzxSaDSXYfjF4wz7XB",
		"secret": k.generateSecret("Ta28hMQmpTA2FCyzxSaDSXYfjF4wz7XB", "Hm_Iuvt_cdb524f42f0cer9b268e4v7y735ewrq2324"),
	}
	//searchCookie, err := k.requester.Get("http://kuwo.cn/search/list?key=any", nil)
	//fmt.Println(searchCookie.Header(), err)
}

func (k *Kuwo) GetName() string {
	return "kuwo"
}

func (k *Kuwo) MatchMedia(keyword string) (miaosic.MediaMeta, bool) {
	if id := k.IdRegex0.FindString(keyword); id != "" {
		return miaosic.MediaMeta{
			Provider:   k.GetName(),
			Identifier: id,
		}, true
	}
	if id := k.IdRegex1.FindString(keyword); id != "" {
		return miaosic.MediaMeta{
			Provider:   k.GetName(),
			Identifier: id[2:],
		}, true
	}
	return miaosic.MediaMeta{}, false
}

//func (k *Kuwo) MatchPlaylist(uri string) *miaosic.Playlist {
//	var id string
//	id = k.PlaylistRegex0.FindString(uri)
//	if id != "" {
//		return &miaosic.Playlist{
//			Meta: miaosic.MediaMeta{k.GetName(), id},
//		}
//	}
//	id = k.PlaylistRegex1.FindString(uri)
//	if id != "" {
//		return &miaosic.Playlist{
//			Meta: miaosic.MediaMeta{k.GetName(), id[9:]},
//		}
//	}
//	return nil
//}

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
			num1 = cast.ToInt64(n[10:11])
			// 19+6+1
			num2 = cast.ToInt64(n[19 : 19+6+1])
			if num2%10 >= 5 {
				num2 = num2/10 + 1
			} else {
				num2 = num2 / 10
			}
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
//						Meta: miaosic.MediaMeta{
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
