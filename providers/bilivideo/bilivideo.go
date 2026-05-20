package bilivideo

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/go-resty/resty/v2"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
)

var _ = (miaosic.MediaProvider)(&BilibiliVideo{})

var biliHeaders = map[string]string{
	"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36",
	"Referer":    "https://www.bilibili.com/",
	"Origin":     "https://www.bilibili.com",
	"Cookie":     "buvid4=86D528BE-0AAA-CF7C-563F-48F8E3F133C140694-023050121-%2FNOyo2MbSyMEJ2aIs%2BzIdw%3D%3D",
}

type BilibiliVideo struct {
	BVRegex   *regexp.Regexp
	IdRegex   *regexp.Regexp
	PageRegex *regexp.Regexp
	header    map[string]string
	client    *resty.Client
}

func NewBilibiliViedo() *BilibiliVideo {
	pvdr := &BilibiliVideo{
		BVRegex:   regexp.MustCompile("^BV[0-9A-Za-z]+"),
		IdRegex:   regexp.MustCompile("^BV[0-9A-Za-z]+(\\?p=[0-9]+)?"),
		PageRegex: regexp.MustCompile("p=[0-9]+"),
		header:    biliHeaders,
	}
	pvdr.client = resty.New().SetTimeout(3 * time.Second)
	//pvdr.FileFunc = buildFileApi(requester, headers)
	//pvdr.PlaylistFunc = buildPlaylistApi(requester, headers)
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

func (b *BilibiliVideo) Search(keyword string, page, size int) ([]miaosic.MediaInfo, error) {
	resp, err := b.client.R().
		SetHeaders(b.header).
		SetQueryParams(map[string]string{
			"keyword":   keyword,
			"page":      strconv.Itoa(page),
			"page_size": strconv.Itoa(size),
		}).
		Get("https://api.bilibili.com/x/web-interface/wbi/search/type?search_type=video")
	if err != nil {
		return nil, err
	}
	respResult := gjson.ParseBytes(resp.Body())
	if respResult.Get("code").String() != "0" {
		return nil, errors.New("failed to find required data")
	}
	//fmt.Println(resp.String())
	result := make([]miaosic.MediaInfo, 0)
	r := regexp.MustCompile("</?em[^>]*>")
	respResult.Get("data.result").ForEach(func(key, value gjson.Result) bool {
		result = append(result, miaosic.MediaInfo{
			Title:   r.ReplaceAllString(value.Get("title").String(), ""),
			Cover:   miaosic.Picture{Url: "https:" + value.Get("pic").String()},
			Artist:  value.Get("author").String(),
			Artists: []string{value.Get("author").String()},
			Meta: miaosic.MetaData{
				Provider:   b.GetName(),
				Identifier: value.Get("bvid").String(),
			},
		})
		return true
	})
	return result, nil
}

func (b *BilibiliVideo) GetMediaInfo(meta miaosic.MetaData) (miaosic.MediaInfo, error) {
	resp, err := b.client.R().
		SetHeaders(b.header).
		SetQueryParam("bvid", b.getBv(meta.Identifier)).
		Get("https://api.bilibili.com/x/web-interface/view/detail?&aid=&jsonp=jsonp")
	if err != nil {
		return miaosic.MediaInfo{}, err
	}
	result := gjson.ParseBytes(resp.Body())
	media := miaosic.MediaInfo{}
	if result.Get("data.View.title").String() == "" {
		return miaosic.MediaInfo{}, errors.New("failed to find required data")
	}
	media.Title = result.Get("data.View.title").String()
	media.Artist = result.Get("data.View.owner.name").String()
	media.Artists = []string{media.Artist}
	media.Cover.Url = result.Get("data.View.pic").String()
	media.Meta = meta
	return media, nil
}

func (b *BilibiliVideo) GetMediaLyric(meta miaosic.MetaData) ([]miaosic.Lyrics, error) {
	return nil, miaosic.ErrNotImplemented
}

func (b *BilibiliVideo) getCids(bvid string) ([]string, error) {
	resp, err := b.client.R().
		SetHeaders(b.header).
		SetQueryParam("bvid", bvid).
		Get("https://api.bilibili.com/x/web-interface/view/detail?&aid=&jsonp=jsonp")
	if err != nil {
		return nil, err
	}
	respResult := gjson.ParseBytes(resp.Body())
	result := make([]string, 0)
	for _, r := range respResult.Get("data.View.pages.#.cid").Array() {
		result = append(result, r.String())
	}
	if len(result) == 0 {
		return nil, errors.New("failed to find cid data")
	}
	return result, nil
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
	cids, err := b.getCids(b.getBv(meta.Identifier))
	if err != nil {
		return nil, err
	}
	if err != nil || page >= len(cids) {
		return nil, miaosic.ErrorExternalApi
	}
	resp, err := b.client.R().
		SetHeaders(b.header).
		SetQueryParams(map[string]string{
			"bvid": b.getBv(meta.Identifier),
			"cid":  cids[page],
			"qn":   string(b.MapQuality(quality)),
		}).
		Get("https://api.bilibili.com/x/player/playurl?type=&otype=json&fourk=1&avid=")
	if err != nil {
		return nil, err
	}
	result := gjson.ParseBytes(resp.Body())
	uri := result.Get("data.durl.0.url").String()
	if uri == "" {
		return nil, miaosic.ErrorExternalApi
	}
	header := make(map[string]string)
	_ = copier.Copy(&header, &b.header)
	header["Referer"] = fmt.Sprintf("https://www.bilibili.com/video/%s", b.getBv(meta.Identifier))
	return []miaosic.MediaUrl{
		{
			Quality: miaosic.QualityStandard,
			Url:     uri,
			Header:  header,
		},
	}, nil
}
