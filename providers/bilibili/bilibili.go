package bilibili

import (
	"github.com/AynaLivePlayer/miaosic"
	"github.com/AynaLivePlayer/miaosic/providers"
	"github.com/aynakeya/deepcolor"
	"github.com/aynakeya/deepcolor/dphttp"
	"github.com/tidwall/gjson"
	"regexp"
)

var _ = (miaosic.MediaProvider)(&Bilibili{})

type Bilibili struct {
	providers.DeepcolorProvider
	IdRegex0 *regexp.Regexp
	IdRegex1 *regexp.Regexp
	header   map[string]string
}

func NewBilibili() *Bilibili {
	bili := &Bilibili{
		IdRegex0: regexp.MustCompile("^[0-9]+"),
		IdRegex1: regexp.MustCompile("^au[0-9]+"),
		header: map[string]string{
			"user-agent": "BiliMusic/2.233.3",
		},
	}
	bili.InfoApi = deepcolor.CreateApiResultFunc(
		func(meta miaosic.MediaMeta) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithSingleQuery(
				"https://www.bilibili.com/audio/music-service-c/web/song/info",
				"sid", meta.Identifier,
				bili.header)
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, media *miaosic.MediaInfo) error {
			if resp.Get("data.title").String() == "" {
				return miaosic.ErrorExternalApi
			}
			media.Title = resp.Get("data.title").String()
			media.Cover.Url = resp.Get("data.cover").String()
			media.Artist = resp.Get("data.author").String()
			return nil
		})
	bili.FileApi = deepcolor.CreateApiResultFunc(
		func(param providers.FileApiParam) (*dphttp.Request, error) {
			// todo: handle quality
			return deepcolor.NewGetRequestWithSingleQuery(
				"https://api.bilibili.com/audio/music-service-c/url?device=phone&mid=8047632&mobi_app=iphone&platform=ios&privilege=2&quality=2",
				"songid", param.Meta.Identifier,
				bili.header)
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, media *[]miaosic.MediaUrl) error {
			if resp.Get("data.cdns.0").String() == "" {
				return miaosic.ErrorExternalApi
			}
			resp.Get("data.cdns").ForEach(func(key, value gjson.Result) bool {
				*media = append(*media, miaosic.NewMediaUrl(value.String(), miaosic.QualityUnk))
				return true
			})
			return nil
		})
	bili.SearchApi = deepcolor.CreateApiResultFunc(
		func(param providers.MediaSearchParam) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithQuery(
				"https://api.bilibili.com/audio/music-service-c/s?search_type=music",
				map[string]any{
					"keyword":  param.Keyword,
					"page":     param.Page,
					"pagesize": param.PageSize,
				},
				bili.header)
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, result *[]miaosic.MediaInfo) error {
			// Assuming data contains a list of search results
			for _, r := range resp.Get("data.result").Array() {
				media := miaosic.MediaInfo{
					Title:  r.Get("title").String(),
					Cover:  miaosic.Picture{Url: r.Get("cover").String()},
					Artist: r.Get("author").String(),
					Meta: miaosic.MediaMeta{
						Provider:   bili.GetName(),
						Identifier: r.Get("id").String(),
					},
				}
				*result = append(*result, media)
			}
			return nil
		})
	return bili
}

func (b *Bilibili) GetName() string {
	return "bilibili"
}

func (b *Bilibili) MatchMedia(keyword string) (miaosic.MediaMeta, bool) {
	if id := b.IdRegex0.FindString(keyword); id != "" {
		return miaosic.MediaMeta{
			Provider:   b.GetName(),
			Identifier: id,
		}, true
	}
	if id := b.IdRegex1.FindString(keyword); id != "" {
		return miaosic.MediaMeta{
			Provider:   b.GetName(),
			Identifier: id[2:],
		}, true
	}
	return miaosic.MediaMeta{}, false
}
