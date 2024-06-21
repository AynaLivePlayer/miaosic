package bilivideo

import (
	"errors"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/AynaLivePlayer/miaosic/providers"
	"github.com/aynakeya/deepcolor"
	"github.com/aynakeya/deepcolor/dphttp"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"regexp"
)

var _ = (miaosic.MediaProvider)(&BilibiliVideo{})

var biliHeaders = map[string]string{
	"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Referer":    "https://www.bilibili.com/",
	"Origin":     "https://www.bilibili.com",
	"Cookie":     "buvid3=40BA0253-7F5C-06C1-12CE-871EC008DB2096426infoc;",
}

type BilibiliVideo struct {
	providers.DeepcolorProvider
	BVRegex   *regexp.Regexp
	IdRegex   *regexp.Regexp
	PageRegex *regexp.Regexp
	header    map[string]string
	cidApi    dphttp.ApiResultFunc[string, []string]
}

func NewBilibiliViedo() *BilibiliVideo {
	pvdr := &BilibiliVideo{
		BVRegex:   regexp.MustCompile("^BV[0-9A-Za-z]+"),
		IdRegex:   regexp.MustCompile("^BV[0-9A-Za-z]+(\\?p=[0-9]+)?"),
		PageRegex: regexp.MustCompile("p=[0-9]+"),
		header:    biliHeaders,
	}
	pvdr.InfoApi = deepcolor.CreateApiResultFunc(
		func(meta miaosic.MetaData) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithQuery(
				"https://api.bilibili.com/x/web-interface/view/detail?&aid=&jsonp=jsonp",
				map[string]any{
					"bvid": pvdr.getBv(meta.Identifier),
				}, pvdr.header,
			)
		},
		deepcolor.ParserGJson,
		func(result *gjson.Result, media *miaosic.MediaInfo) error {
			if result.Get("data.View.title").String() == "" {
				return errors.New("failed to find required data")
			}
			media.Title = result.Get("data.View.title").String()
			media.Artist = result.Get("data.View.owner.name").String()
			media.Cover.Url = result.Get("data.View.pic").String()
			return nil
		})

	pvdr.cidApi = deepcolor.CreateApiResultFunc(
		deepcolor.NewGetRequestFuncWithSingleQuery(
			"https://api.bilibili.com/x/web-interface/view/detail?&aid=&jsonp=jsonp", "bvid", pvdr.header),
		deepcolor.ParserGJson,
		func(resp *gjson.Result, result *[]string) error {
			for _, r := range resp.Get("data.View.pages.#.cid").Array() {
				*result = append(*result, r.String())
			}
			if len(*result) == 0 {
				return errors.New("failed to find cid data")
			}
			return nil
		})
	//pvdr.FileFunc = buildFileApi(requester, headers)
	//pvdr.PlaylistFunc = buildPlaylistApi(requester, headers)
	pvdr.SearchApi = deepcolor.CreateApiResultFunc(
		func(param providers.MediaSearchParam) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithQuery(
				"https://api.bilibili.com/x/web-interface/wbi/search/type?search_type=video",
				map[string]any{
					"keyword":   param.Keyword,
					"page":      param.Page,
					"page_size": param.PageSize,
				}, pvdr.header)
		},
		deepcolor.ParserGJson,
		func(resp *gjson.Result, result *[]miaosic.MediaInfo) error {
			if resp.Get("code").String() != "0" {
				return errors.New("failed to find required data")
			}
			r := regexp.MustCompile("</?em[^>]*>")
			resp.Get("data.result").ForEach(func(key, value gjson.Result) bool {
				*result = append(*result, miaosic.MediaInfo{
					Title:  r.ReplaceAllString(value.Get("title").String(), ""),
					Cover:  miaosic.Picture{Url: "https:" + value.Get("pic").String()},
					Artist: value.Get("author").String(),
					Meta: miaosic.MetaData{
						Provider:   pvdr.GetName(),
						Identifier: value.Get("bvid").String(),
					},
				})
				return true
			})
			return nil
		})
	return pvdr
}

func (b *BilibiliVideo) getPage(bv string) int {
	if page := b.PageRegex.FindString(bv); page != "" {
		return cast.ToInt(page[2:])
	}
	return 1
}

func (b *BilibiliVideo) getBv(bv string) string {
	return b.BVRegex.FindString(bv)
}

func (b *BilibiliVideo) GetName() string {
	return "bilibili-video"
}

func (b *BilibiliVideo) MatchMedia(keyword string) (miaosic.MetaData, bool) {
	if id := b.IdRegex.FindString(keyword); id != "" {
		return miaosic.MetaData{
			Provider:   b.GetName(),
			Identifier: id,
		}, true
	}
	return miaosic.MetaData{}, false
}

func (b *BilibiliVideo) GetMediaUrl(meta miaosic.MetaData, quality miaosic.Quality) ([]miaosic.MediaUrl, error) {
	page := b.getPage(meta.Identifier) - 1
	cids, err := b.cidApi(b.getBv(meta.Identifier))
	if err != nil {
		return nil, err
	}
	if err != nil || page >= len(cids) {
		return nil, miaosic.ErrorExternalApi
	}
	return deepcolor.CreateApiResultFunc(
		func(param string) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithQuery(
				"https://api.bilibili.com/x/player/playurl?type=&otype=json&fourk=1&qn=32&avid=",
				map[string]any{
					"bvid": b.getBv(meta.Identifier),
					"cid":  cids[page],
				}, b.header)
		},
		deepcolor.ParserGJson,
		func(result *gjson.Result, container *[]miaosic.MediaUrl) error {
			uri := result.Get("data.durl.0.url").String()
			if uri == "" {
				return miaosic.ErrorExternalApi
			}
			header := make(map[string]string)
			_ = copier.Copy(&header, &b.header)
			header["Referer"] = fmt.Sprintf("https://www.bilibili.com/video/%s", b.getBv(meta.Identifier))
			*container = append(*container, miaosic.MediaUrl{
				Quality: miaosic.QualityUnk,
				Url:     uri,
				Header:  header,
			})
			return nil
		})("")
}
