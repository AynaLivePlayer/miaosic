package bilibili

import (
	"github.com/aynakeya/deepcolor"
	"github.com/aynakeya/deepcolor/dphttp"
	"github.com/tidwall/gjson"
	"miaosic"
	"miaosic/providers"
	"regexp"
)

type Bilibili struct {
	providers.DeepcolorProvider
	requester dphttp.IRequester
	IdRegex0  *regexp.Regexp
	IdRegex1  *regexp.Regexp
	header    map[string]string
}

func NewBilibili(requester dphttp.IRequester) *Bilibili {
	bili := &Bilibili{
		requester: requester,
		IdRegex0:  regexp.MustCompile("^[0-9]+"),
		IdRegex1:  regexp.MustCompile("^au[0-9]+"),
	}
	bili.initHeader()
	bili.InfoFunc = bili.buildInfoApi()
	bili.FileFunc = bili.buildFileApi()
	bili.SearchFunc = bili.buildSearchApi()
	return bili
}

func (b *Bilibili) GetName() string {
	return "bilibili"
}

func (b *Bilibili) MatchMedia(keyword string) *miaosic.Media {
	if id := b.IdRegex0.FindString(keyword); id != "" {
		return &miaosic.Media{
			Meta: miaosic.MediaMeta{
				Provider:   b.GetName(),
				Identifier: id,
			},
		}
	}
	if id := b.IdRegex1.FindString(keyword); id != "" {
		return &miaosic.Media{
			Meta: miaosic.MediaMeta{
				Provider:   b.GetName(),
				Identifier: id[2:],
			},
		}
	}
	return nil
}

func (b *Bilibili) MatchPlaylist(keyword string) *miaosic.Playlist {
	return nil
}

func (b *Bilibili) initHeader() {
	b.header = map[string]string{
		"user-agent": "BiliMusic/2.233.3",
	}
}

func (b *Bilibili) buildInfoApi() dphttp.ApiFunc[*miaosic.Media, *miaosic.Media] {
	return deepcolor.CreateApiFunc(
		b.requester,
		func(media *miaosic.Media) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithSingleQuery("https://www.bilibili.com/audio/music-service-c/web/song/info", "sid", b.header)(media.Meta.Identifier)
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, media *miaosic.Media) error {
			if resp.Get("data.title").String() == "" {
				return miaosic.ErrorExternalApi
			}
			media.Title = resp.Get("data.title").String()
			media.Cover.Url = resp.Get("data.cover").String()
			media.Artist = resp.Get("data.author").String()
			return nil
		})
}

func (b *Bilibili) buildFileApi() dphttp.ApiFunc[*miaosic.Media, *miaosic.Media] {
	return deepcolor.CreateApiFunc(
		b.requester,
		func(media *miaosic.Media) (*dphttp.Request, error) {
			// Assuming the endpoint and query are similar for file details
			return deepcolor.NewGetRequestWithSingleQuery("https://api.bilibili.com/audio/music-service-c/url?device=phone&mid=8047632&mobi_app=iphone&platform=ios&privilege=2&quality=2", "songid", b.header)(media.Meta.Identifier)
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, media *miaosic.Media) error {
			if resp.Get("data.cdns.0").String() == "" {
				return miaosic.ErrorExternalApi
			}
			media.Url = resp.Get("data.cdns.0").String()
			return nil
		})
}

func (b *Bilibili) buildSearchApi() dphttp.ApiFuncResult[string, []*miaosic.Media] {
	return deepcolor.CreateApiResultFunc(
		b.requester,
		func(query string) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithSingleQuery("https://api.bilibili.com/audio/music-service-c/s?search_type=music&page=1&pagesize=100", "keyword", b.header)(query)
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, result *[]*miaosic.Media) error {
			// Assuming data contains a list of search results
			for _, r := range resp.Get("data.result").Array() {
				media := &miaosic.Media{
					Title:  r.Get("title").String(),
					Cover:  miaosic.Picture{Url: r.Get("cover").String()},
					Artist: r.Get("author").String(),
					Meta: miaosic.MediaMeta{
						Provider:   b.GetName(),
						Identifier: r.Get("id").String(),
					},
				}
				*result = append(*result, media)
			}
			return nil
		})
}
