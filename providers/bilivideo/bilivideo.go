package bilivideo

import (
	"errors"
	"fmt"
	"github.com/aynakeya/deepcolor"
	"github.com/aynakeya/deepcolor/dphttp"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"miaosic"
	"miaosic/providers"
	"regexp"
)

var _ = (miaosic.MediaProvider)(&BilibiliVideo{})

type BilibiliVideo struct {
	providers.DeepcolorProvider
	requester dphttp.IRequester
	BVRegex   *regexp.Regexp
	IdRegex   *regexp.Regexp
	PageRegex *regexp.Regexp
	header    map[string]string
}

func NewBilibiliViedo(requester dphttp.IRequester) *BilibiliVideo {
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:51.0) Gecko/20100101 Firefox/51.0",
		"Referer":    "https://www.bilibili.com/",
		"Origin":     "https://www.bilibili.com",
		"Cookie":     "buvid3=9A8B3564-BDA9-407F-B45F-D5C40786CA49167618infoc;",
	}
	pvdr := &BilibiliVideo{
		requester: requester,
		BVRegex:   regexp.MustCompile("^BV[0-9A-Za-z]+"),
		IdRegex:   regexp.MustCompile("^BV[0-9A-Za-z]+(\\?p=[0-9]+)?"),
		PageRegex: regexp.MustCompile("p=[0-9]+"),
		header:    headers,
	}
	pvdr.InfoFunc = pvdr.buildInfoApi()
	//pvdr.FileFunc = buildFileApi(requester, headers)
	pvdr.LyricFunc = nil
	//pvdr.PlaylistFunc = buildPlaylistApi(requester, headers)
	pvdr.SearchFunc = pvdr.buildSearchApi()
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

func (b *BilibiliVideo) MatchMedia(keyword string) *miaosic.Media {
	if id := b.IdRegex.FindString(keyword); id != "" {
		return &miaosic.Media{
			Meta: miaosic.MediaMeta{
				Provider:   b.GetName(),
				Identifier: id,
			},
		}
	}
	return nil
}

func (b *BilibiliVideo) MatchPlaylist(keyword string) *miaosic.Playlist {
	return nil
}

func (b *BilibiliVideo) UpdateMediaLyric(media *miaosic.Media) error {
	return nil
}

func (b *BilibiliVideo) buildInfoApi() dphttp.ApiFunc[*miaosic.Media, *miaosic.Media] {
	return deepcolor.CreateApiFunc(
		b.requester,
		func(params *miaosic.Media) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithSingleQuery(
				"https://api.bilibili.com/x/web-interface/view/detail?&aid=&jsonp=jsonp",
				"bvid", b.header,
			)(b.getBv(params.Meta.Identifier))
		},
		deepcolor.ParserGJson,
		func(result *gjson.Result, media *miaosic.Media) error {
			if result.Get("data.View.title").String() == "" {
				return errors.New("failed to find required data")
			}
			media.Title = result.Get("data.View.title").String()
			media.Artist = result.Get("data.View.owner.name").String()
			media.Cover.Url = result.Get("data.View.pic").String()
			return nil
		})
}

func (b *BilibiliVideo) buildSearchApi() dphttp.ApiFuncResult[string, []*miaosic.Media] {
	return deepcolor.CreateApiResultFunc(
		b.requester,
		deepcolor.NewGetRequestWithSingleQuery(
			"https://api.bilibili.com/x/web-interface/wbi/search/type?search_type=video&page=1",
			"keyword", b.header),
		deepcolor.ParserGJson,
		func(resp *gjson.Result, result *[]*miaosic.Media) error {
			if resp.Get("code").String() != "0" {
				return errors.New("failed to find required data")
			}
			r := regexp.MustCompile("</?em[^>]*>")
			resp.Get("data.result").ForEach(func(key, value gjson.Result) bool {
				*result = append(*result, &miaosic.Media{
					Title:  r.ReplaceAllString(value.Get("title").String(), ""),
					Cover:  miaosic.Picture{Url: "https:" + value.Get("pic").String()},
					Artist: value.Get("author").String(),
					Meta: miaosic.MediaMeta{
						Provider:   b.GetName(),
						Identifier: value.Get("bvid").String(),
					},
				})
				return true
			})
			return nil
		})
}

func (b *BilibiliVideo) cidApi(bvid string) ([]string, error) {
	return deepcolor.CreateApiResultFunc(
		b.requester,
		deepcolor.NewGetRequestWithSingleQuery(
			"https://api.bilibili.com/x/web-interface/view/detail?&aid=&jsonp=jsonp", "bvid", b.header),
		deepcolor.ParserGJson,
		func(resp *gjson.Result, result *[]string) error {
			for _, r := range resp.Get("data.View.pages.#.cid").Array() {
				*result = append(*result, r.String())
			}
			if len(*result) == 0 {
				return errors.New("failed to find cid data")
			}
			return nil
		})(bvid)
}

func (b *BilibiliVideo) UpdateMediaUrl(media *miaosic.Media) error {
	page := b.getPage(media.Meta.Identifier) - 1
	cids, err := b.cidApi(b.getBv(media.Meta.Identifier))
	if err != nil {
		return err
	}
	if err != nil || page >= len(cids) {
		return miaosic.ErrorExternalApi
	}
	return deepcolor.CreateApiFunc(
		b.requester,
		func(params *miaosic.Media) (*dphttp.Request, error) {
			return deepcolor.NewGetRequestWithQuery(
				"https://api.bilibili.com/x/player/playurl?type=&otype=json&fourk=1&qn=32&avid=",
				[]string{"bvid", "cid"}, b.header,
			)([]string{b.getBv(media.Meta.Identifier), cids[page]})
		},
		deepcolor.ParserGJson,
		func(result *gjson.Result, container *miaosic.Media) error {
			uri := result.Get("data.durl.0.url").String()
			if uri == "" {
				return miaosic.ErrorExternalApi
			}
			container.Url = uri
			header := make(map[string]string)
			_ = copier.Copy(&header, &b.header)
			header["Referer"] = fmt.Sprintf("https://www.bilibili.com/video/%s", b.getBv(container.Meta.Identifier))
			container.Header = header
			return nil
		})(media, media)
}
