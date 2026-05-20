package bilibili

import (
	"regexp"
	"strconv"
	"time"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

var _ = (miaosic.MediaProvider)(&Bilibili{})

type Bilibili struct {
	IdRegex0 *regexp.Regexp
	IdRegex1 *regexp.Regexp
	header   map[string]string
	client   *resty.Client
}

func NewBilibili() *Bilibili {
	bili := &Bilibili{
		IdRegex0: regexp.MustCompile("^[0-9]+"),
		IdRegex1: regexp.MustCompile("^au[0-9]+"),
		header: map[string]string{
			"user-agent": "BiliMusic/2.233.3",
		},
	}
	bili.client = resty.New().SetTimeout(3 * time.Second)
	return bili
}

/*
	bili.InfoApi = deepcolor.CreateApiResultFunc(
		func(meta miaosic.MetaData) (*dphttp.Request, error) {
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
					Meta: miaosic.MetaData{
						Provider:   bili.GetName(),
						Identifier: r.Get("id").String(),
					},
				}
				*result = append(*result, media)
			}
			return nil
		})
*/

func (b *Bilibili) GetName() string {
	return "bilibili"
}

func (b *Bilibili) Search(keyword string, page, size int) ([]miaosic.MediaInfo, error) {
	resp, err := b.client.R().
		SetHeaders(b.header).
		SetQueryParams(map[string]string{
			"keyword":  keyword,
			"page":     strconv.Itoa(page),
			"pagesize": strconv.Itoa(size),
		}).
		Get("https://api.bilibili.com/audio/music-service-c/s?search_type=music")
	if err != nil {
		return nil, err
	}
	respResult := gjson.ParseBytes(resp.Body())
	result := make([]miaosic.MediaInfo, 0)
	// Assuming data contains a list of search results
	for _, r := range respResult.Get("data.result").Array() {
		media := miaosic.MediaInfo{
			Title:  r.Get("title").String(),
			Cover:  miaosic.Picture{Url: r.Get("cover").String()},
			Artist: r.Get("author").String(),
			Meta: miaosic.MetaData{
				Provider:   b.GetName(),
				Identifier: r.Get("id").String(),
			},
		}
		result = append(result, media)
	}
	return result, nil
}

func (b *Bilibili) GetMediaInfo(meta miaosic.MetaData) (miaosic.MediaInfo, error) {
	resp, err := b.client.R().
		SetHeaders(b.header).
		SetQueryParam("sid", meta.Identifier).
		Get("https://www.bilibili.com/audio/music-service-c/web/song/info")
	if err != nil {
		return miaosic.MediaInfo{}, err
	}
	result := gjson.ParseBytes(resp.Body())
	media := miaosic.MediaInfo{}
	if result.Get("data.title").String() == "" {
		return miaosic.MediaInfo{}, miaosic.ErrorExternalApi
	}
	media.Title = result.Get("data.title").String()
	media.Cover.Url = result.Get("data.cover").String()
	media.Artist = result.Get("data.author").String()
	media.Meta = meta
	return media, nil
}

func (b *Bilibili) GetMediaUrl(meta miaosic.MetaData, quality miaosic.Quality) ([]miaosic.MediaUrl, error) {
	// todo: handle quality
	resp, err := b.client.R().
		SetHeaders(b.header).
		SetQueryParam("songid", meta.Identifier).
		SetQueryParam("quality", string(b.MapQuality(quality))).
		Get("https://api.bilibili.com/audio/music-service-c/url?device=phone&mid=8047632&mobi_app=iphone&platform=ios&privilege=2")
	if err != nil {
		return nil, err
	}
	result := gjson.ParseBytes(resp.Body())
	if result.Get("data.cdns.0").String() == "" {
		return nil, miaosic.ErrorExternalApi
	}
	media := make([]miaosic.MediaUrl, 0)
	result.Get("data.cdns").ForEach(func(key, value gjson.Result) bool {
		media = append(media, miaosic.NewMediaUrl(value.String(), miaosic.QualityStandard))
		return true
	})
	return media, nil
}

func (b *Bilibili) GetMediaLyric(meta miaosic.MetaData) ([]miaosic.Lyrics, error) {
	return nil, miaosic.ErrNotImplemented
}

func (b *Bilibili) MatchMedia(keyword string) (miaosic.MetaData, bool) {
	if id := b.IdRegex0.FindString(keyword); id != "" {
		return miaosic.MetaData{
			Provider:   b.GetName(),
			Identifier: id,
		}, true
	}
	if id := b.IdRegex1.FindString(keyword); id != "" {
		return miaosic.MetaData{
			Provider:   b.GetName(),
			Identifier: id[2:],
		}, true
	}
	return miaosic.MetaData{}, false
}
