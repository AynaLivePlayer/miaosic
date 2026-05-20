package migu

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/AynaLivePlayer/miaosic/utils"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

const (
	userAgent   = "Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1"
	referer     = "http://music.migu.cn/"
	magicUserID = "15548614588710179085069"
)

type Migu struct {
	cookie           string
	client           *resty.Client
	noRedirectClient *resty.Client
	idRegex          *regexp.Regexp
	linkRegexes      []*regexp.Regexp
}

func NewMigu() *Migu {
	return NewMiguWithCookie("")
}

func NewMiguWithCookie(cookie string) *Migu {
	return &Migu{
		cookie:           cookie,
		client:           newMiguClient(),
		noRedirectClient: newMiguClient().SetRedirectPolicy(resty.NoRedirectPolicy()),
		idRegex:          regexp.MustCompile(`^(?:mg)?\d+$`),
		linkRegexes: []*regexp.Regexp{
			regexp.MustCompile(`music\.migu\.cn/(?:v3|v5)/music/song/(\d+)`),
			regexp.MustCompile(`contentId=(\d+)`),
			regexp.MustCompile(`copyrightId=(\d+)`),
		},
	}
}

func newMiguClient() *resty.Client {
	return resty.New().
		SetTimeout(8 * time.Second).
		SetRetryCount(2).
		SetRetryWaitTime(300 * time.Millisecond).
		SetRetryMaxWaitTime(2 * time.Second)
}

func (m *Migu) GetName() string {
	return "migu"
}

func (m *Migu) requestHeaders() map[string]string {
	return map[string]string{
		"User-Agent": userAgent,
		"Referer":    referer,
		"Cookie":     m.cookie,
	}
}

func (m *Migu) Search(keyword string, page, size int) ([]miaosic.MediaInfo, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	const apiPageSize = 20
	offset := (page - 1) * size
	apiPage := offset/apiPageSize + 1
	start := offset % apiPageSize

	medias := make([]miaosic.MediaInfo, 0, size)
	for len(medias) < size {
		apiMedias, err := m.searchPage(keyword, apiPage)
		if err != nil {
			return nil, err
		}
		if len(apiMedias) == 0 {
			break
		}
		if start >= len(apiMedias) {
			break
		}

		for _, media := range apiMedias[start:] {
			medias = append(medias, media)
			if len(medias) >= size {
				break
			}
		}
		if len(apiMedias) < apiPageSize {
			break
		}
		apiPage++
		start = 0
	}
	return medias, nil
}

func (m *Migu) searchPage(keyword string, page int) ([]miaosic.MediaInfo, error) {
	resp, err := m.client.R().
		SetHeaders(m.requestHeaders()).
		SetQueryParams(map[string]string{
			"ua":           "Android_migu",
			"version":      "5.0.1",
			"text":         keyword,
			"pageNo":       strconv.Itoa(page),
			"pageSize":     "20",
			"searchSwitch": `{"song":1,"album":0,"singer":0,"tagSong":0,"mvSong":0,"songlist":0,"bestShow":1}`,
		}).
		Get("http://pd.musicapp.migu.cn/MIGUM2.0/v1.0/content/search_all.do")
	if err != nil {
		return nil, err
	}

	var result struct {
		SongResultData struct {
			Result []miguSongItem `json:"result"`
		} `json:"songResultData"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("migu: search json parse error: %w", err)
	}

	medias := make([]miaosic.MediaInfo, 0, len(result.SongResultData.Result))
	for _, item := range result.SongResultData.Result {
		media, ok := m.convertItemToMedia(item, miaosic.QualityAny, true)
		if ok {
			medias = append(medias, media)
		}
	}
	return medias, nil
}

func (m *Migu) MatchMedia(uri string) (miaosic.MetaData, bool) {
	for _, pattern := range m.linkRegexes {
		matches := pattern.FindStringSubmatch(uri)
		if len(matches) >= 2 {
			return miaosic.NewMetaData(m.GetName(), matches[1]), true
		}
	}
	if id := m.idRegex.FindString(uri); id != "" {
		return miaosic.NewMetaData(m.GetName(), strings.TrimPrefix(id, "mg")), true
	}
	return miaosic.MetaData{}, false
}

func (m *Migu) GetMediaInfo(meta miaosic.MetaData) (miaosic.MediaInfo, error) {
	contentID, _, formatType := splitMiguIdentifier(meta.Identifier)
	if contentID == "" {
		return miaosic.MediaInfo{}, miaosic.ErrorInvalidMediaMeta
	}
	item, err := m.fetchSongDetail(contentID)
	if err != nil {
		return miaosic.MediaInfo{}, err
	}
	media, ok := m.convertItemToMedia(item, miaosic.Quality(formatType), true)
	if !ok {
		return miaosic.MediaInfo{}, miaosic.ErrorExternalApi
	}
	return media, nil
}

func (m *Migu) GetMediaUrl(meta miaosic.MetaData, quality miaosic.Quality) ([]miaosic.MediaUrl, error) {
	contentID, resourceType, formatType := splitMiguIdentifier(meta.Identifier)
	if contentID == "" {
		return nil, miaosic.ErrorInvalidMediaMeta
	}

	requestedQuality := m.MapQuality(quality)
	if resourceType == "" || formatType == "" || requestedQuality != miaosic.Quality(formatType) {
		item, err := m.fetchSongDetail(contentID)
		if err != nil {
			return nil, err
		}
		format, ok := pickMiguFormat(item.rateFormats(), requestedQuality, true)
		if !ok {
			return nil, miaosic.ErrorExternalApi
		}
		resourceType = format.ResourceType
		formatType = format.FormatType
	}

	resp, err := m.noRedirectClient.R().
		SetHeaders(m.requestHeaders()).
		SetQueryParams(map[string]string{
			"toneFlag":     formatType,
			"netType":      "00",
			"userId":       magicUserID,
			"ua":           "Android_migu",
			"version":      "5.1",
			"copyrightId":  "0",
			"contentId":    contentID,
			"resourceType": resourceType,
			"channel":      "0",
		}).
		Get("http://app.pd.nf.migu.cn/MIGUM2.0/v1.0/content/sub/listenSong.do")
	if err != nil && resp == nil {
		return nil, err
	}

	// ref/music-lib used to expect listenSong.do to return a 302 Location.
	// The current endpoint may return HTTP 200 with a business-error JSON body,
	// so only explicit redirect/URL fields are treated as playable URLs.
	location := ""
	if resp != nil && resp.RawResponse != nil {
		location = resp.RawResponse.Header.Get("Location")
	}
	if location != "" {
		return []miaosic.MediaUrl{miaosic.NewMediaUrl(location, miaosic.Quality(formatType))}, nil
	}

	body := gjson.ParseBytes(resp.Body())
	for _, path := range []string{"data.url", "data.playUrl", "data.song.url", "url", "playUrl"} {
		if url := body.Get(path).String(); url != "" {
			return []miaosic.MediaUrl{miaosic.NewMediaUrl(url, miaosic.Quality(formatType))}, nil
		}
	}
	if code := body.Get("code").String(); code != "" {
		return nil, fmt.Errorf("migu: failed to get media url: %s (code %s)", body.Get("info").String(), code)
	}
	return nil, miaosic.ErrorExternalApi
}

func (m *Migu) GetMediaLyric(meta miaosic.MetaData) ([]miaosic.Lyrics, error) {
	contentID, _, _ := splitMiguIdentifier(meta.Identifier)
	if contentID == "" {
		return nil, miaosic.ErrorInvalidMediaMeta
	}

	resp, err := m.client.R().
		SetHeaders(m.requestHeaders()).
		SetQueryParams(map[string]string{
			"resourceId":   contentID,
			"resourceType": "2",
		}).
		Get("http://c.musicapp.migu.cn/MIGUM2.0/v1.0/content/resourceinfo.do")
	if err != nil {
		return nil, err
	}

	var result struct {
		Resource []struct {
			LrcURL   string `json:"lrcUrl"`
			LyricURL string `json:"lyricUrl"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("migu: lyric resource json parse error: %w", err)
	}
	if len(result.Resource) == 0 {
		return nil, errors.New("migu: lyric resource not found")
	}
	lyricURL := firstNonEmpty(result.Resource[0].LrcURL, result.Resource[0].LyricURL)
	if lyricURL == "" {
		return nil, errors.New("migu: lyric url not found")
	}
	lyricURL = strings.Replace(lyricURL, "http://", "https://", 1)

	lyricResp, err := m.client.R().
		SetHeaders(map[string]string{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36",
			"Referer":    "https://y.migu.cn/",
			"Cookie":     m.cookie,
		}).
		Get(lyricURL)
	if err != nil {
		return nil, err
	}

	return []miaosic.Lyrics{utils.ParseLyricWithLangDetection(string(lyricResp.Body()))}, nil
}

func (m *Migu) fetchSongDetail(contentID string) (miguSongItem, error) {
	resp, err := m.client.R().
		SetHeaders(m.requestHeaders()).
		SetQueryParams(map[string]string{
			"resourceId":   contentID,
			"resourceType": "2",
		}).
		Get("http://c.musicapp.migu.cn/MIGUM2.0/v1.0/content/resourceinfo.do")
	if err != nil {
		return miguSongItem{}, err
	}

	var result struct {
		Resource []miguSongItem `json:"resource"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return miguSongItem{}, fmt.Errorf("migu: song detail json parse error: %w", err)
	}
	if len(result.Resource) > 0 {
		return result.Resource[0], nil
	}
	return miguSongItem{}, errors.New("migu: song detail not found")
}

func (m *Migu) convertItemToMedia(item miguSongItem, quality miaosic.Quality, allowPaid bool) (miaosic.MediaInfo, bool) {
	format, ok := pickMiguFormat(item.rateFormats(), m.MapQuality(quality), allowPaid)
	if !ok {
		return miaosic.MediaInfo{}, false
	}
	contentID := firstNonEmpty(item.ContentID, item.CopyrightID, item.ID, item.SongID)
	if contentID == "" {
		return miaosic.MediaInfo{}, false
	}

	artists := collectMiguArtistNames(item)
	albumName := strings.TrimSpace(item.Album)
	if len(item.Albums) > 0 && strings.TrimSpace(item.Albums[0].Name) != "" {
		albumName = strings.TrimSpace(item.Albums[0].Name)
	}

	return miaosic.MediaInfo{
		Title:   firstNonEmpty(item.Name, item.SongName),
		Artist:  strings.Join(artists, ","),
		Artists: artists,
		Cover:   miaosic.Picture{Url: normalizeMiguImageURL(firstNonEmpty(pickMiguImage(item.ImgItems), pickMiguImage(item.AlbumImgs), item.Img1, item.Img2, item.Img3))},
		Album:   albumName,
		Meta: miaosic.MetaData{
			Provider:   m.GetName(),
			Identifier: buildMiguIdentifier(contentID, format.ResourceType, format.FormatType),
		},
	}, true
}

type miguImageItem struct {
	ImgSizeType string `json:"imgSizeType"`
	Img         string `json:"img"`
}

type miguArtistItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type miguAlbumItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type miguRateFormat struct {
	FormatType      string   `json:"formatType"`
	ResourceType    string   `json:"resourceType"`
	Size            string   `json:"size"`
	AndroidSize     string   `json:"androidSize"`
	ISize           string   `json:"isize"`
	ASize           string   `json:"asize"`
	FileType        string   `json:"fileType"`
	AndroidFileType string   `json:"androidFileType"`
	IFormat         string   `json:"iformat"`
	AFormat         string   `json:"aformat"`
	Price           string   `json:"price"`
	ShowTag         []string `json:"showTag"`
	ShowTags        []string `json:"showTags"`
}

type miguSongItem struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	SongName        string           `json:"songName"`
	SongID          string           `json:"songId"`
	Singers         []miguArtistItem `json:"singers"`
	Artists         []miguArtistItem `json:"artists"`
	SingerList      []miguArtistItem `json:"singerList"`
	Albums          []miguAlbumItem  `json:"albums"`
	AlbumID         string           `json:"albumId"`
	Album           string           `json:"album"`
	Singer          string           `json:"singer"`
	ContentID       string           `json:"contentId"`
	CopyrightID     string           `json:"copyrightId"`
	ChargeAuditions string           `json:"chargeAuditions"`
	ImgItems        []miguImageItem  `json:"imgItems"`
	AlbumImgs       []miguImageItem  `json:"albumImgs"`
	RateFormats     []miguRateFormat `json:"rateFormats"`
	AudioFormats    []miguRateFormat `json:"audioFormats"`
	Img1            string           `json:"img1"`
	Img2            string           `json:"img2"`
	Img3            string           `json:"img3"`
	Duration        int              `json:"duration"`
}

func (item miguSongItem) rateFormats() []miguRateFormat {
	if len(item.RateFormats) > 0 {
		return item.RateFormats
	}
	return item.AudioFormats
}

func buildMiguIdentifier(contentID, resourceType, formatType string) string {
	return fmt.Sprintf("%s|%s|%s", contentID, resourceType, formatType)
}

func splitMiguIdentifier(identifier string) (contentID, resourceType, formatType string) {
	parts := strings.Split(identifier, "|")
	if len(parts) >= 1 {
		contentID = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		resourceType = strings.TrimSpace(parts[1])
	}
	if len(parts) >= 3 {
		formatType = strings.TrimSpace(parts[2])
	}
	return contentID, resourceType, formatType
}

func collectMiguArtistNames(item miguSongItem) []string {
	names := make([]string, 0, len(item.Singers)+len(item.Artists)+1)
	seen := make(map[string]struct{})

	appendName := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}

	for _, singer := range item.Singers {
		appendName(singer.Name)
	}
	for _, singer := range item.SingerList {
		appendName(singer.Name)
	}
	for _, artist := range item.Artists {
		appendName(artist.Name)
	}
	if len(names) == 0 {
		for _, name := range strings.Split(item.Singer, "|") {
			appendName(name)
		}
	}
	return names
}

func pickMiguImage(items []miguImageItem) string {
	for _, preferred := range []string{"02", "01", "03"} {
		for _, item := range items {
			if item.ImgSizeType == preferred && strings.TrimSpace(item.Img) != "" {
				return item.Img
			}
		}
	}
	for _, item := range items {
		if strings.TrimSpace(item.Img) != "" {
			return item.Img
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func firstNonZeroString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && value != "0" {
			return value
		}
	}
	return ""
}

func normalizeMiguImageURL(image string) string {
	image = strings.TrimSpace(image)
	if image == "" {
		return ""
	}
	if strings.HasPrefix(image, "http://") || strings.HasPrefix(image, "https://") {
		return image
	}
	if strings.HasPrefix(image, "//") {
		return "https:" + image
	}
	if strings.HasPrefix(image, "/") {
		return "https://d.musicapp.migu.cn" + image
	}
	return image
}
