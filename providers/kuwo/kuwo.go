package kuwo

import (
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"math"
	"math/rand"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/AynaLivePlayer/miaosic/utils"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
)

const kuwoConvertURL2Source = "kwplayer_ar_5.1.0.0_B_jiakong_vh.apk"

var kuwoDownloadURLRegex = regexp.MustCompile(`http[^\s"]+`)

type Kuwo struct {
	PlaylistRegex0 *regexp.Regexp
	PlaylistRegex1 *regexp.Regexp
	IdRegex0       *regexp.Regexp
	IdRegex1       *regexp.Regexp
	header         map[string]string
	client         *resty.Client
}

func NewKuwo() *Kuwo {
	kw := &Kuwo{
		PlaylistRegex0: regexp.MustCompile("[0-9]+"),
		PlaylistRegex1: regexp.MustCompile("playlist/[0-9]+"),
		IdRegex0:       regexp.MustCompile("^[0-9]+$"),
		IdRegex1:       regexp.MustCompile("^kw[0-9]+$"),
		client:         resty.New().SetTimeout(3 * time.Second),
	}
	kw.initToken()
	//kw.PlaylistFunc = kw.playlistApi
	return kw
}

func (k *Kuwo) initToken() {
	k.header = map[string]string{
		"accept": "application/json, text/plain, */*",
		"cookie": "Hm_Iuvt_cdb524f42f23cer9b268564v7y735ewrq2324=T3WtFh7AT3ZMkFrrhEGe8iRhA85SdM8b",
		"secret": k.generateSecret("T3WtFh7AT3ZMkFrrhEGe8iRhA85SdM8b", "Hm_Iuvt_cdb524f42f23cer9b268564v7y735ewrq2324"),
	}
	//searchCookie, err := k.requester.Get("http://kuwo.cn/search/list?key=any", nil)
	//fmt.Println(searchCookie.Header(), err)
}

func (k *Kuwo) GetName() string {
	return "kuwo"
}

func (k *Kuwo) Search(keyword string, page, size int) ([]miaosic.MediaInfo, error) {
	resp, err := k.client.R().
		SetHeaders(k.header).
		SetQueryParams(map[string]string{
			"all": keyword,
			"pn":  strconv.Itoa(page - 1),
			"rn":  strconv.Itoa(size),
		}).
		Get("http://www.kuwo.cn/search/searchMusicBykeyWord/searchMusicBykeyWord?vipver=1&client=kt&ft=music&cluster=0&strategy=2012&encoding=utf8&rformat=json&mobi=1&issubtitle=1&show_copyright_off=1")
	if err != nil {
		return nil, err
	}
	respResult := gjson.ParseBytes(resp.Body())
	result := make([]miaosic.MediaInfo, 0)
	respResult.Get("abslist").ForEach(func(key, value gjson.Result) bool {
		result = append(result, miaosic.MediaInfo{
			Title:  html.UnescapeString(value.Get("SONGNAME").String()),
			Cover:  miaosic.Picture{Url: "https://img2.kuwo.cn/star/albumcover/" + value.Get("web_albumpic_short").String()},
			Artist: value.Get("ARTIST").String(),
			Album:  value.Get("ALBUM").String(),
			Meta: miaosic.MetaData{
				Provider:   k.GetName(),
				Identifier: value.Get("DC_TARGETID").String(),
			},
		})
		return true
	})
	return result, nil
}

func (k *Kuwo) GetMediaInfo(meta miaosic.MetaData) (miaosic.MediaInfo, error) {
	resp, err := k.client.R().
		SetHeaders(k.header).
		SetQueryParam("mid", meta.Identifier).
		Get("http://www.kuwo.cn/api/www/music/musicInfo?httpsStatus=1")
	if err != nil {
		return miaosic.MediaInfo{}, err
	}
	respResult := gjson.ParseBytes(resp.Body())
	media := miaosic.MediaInfo{Meta: meta}
	if respResult.Get("data.musicrid").String() == "" {
		return miaosic.MediaInfo{}, miaosic.ErrorExternalApi
	}
	media.Title = html.UnescapeString(respResult.Get("data.name").String())
	media.Cover.Url = respResult.Get("data.pic").String()
	media.Artist = respResult.Get("data.artist").String()
	media.Album = respResult.Get("data.album").String()
	return media, nil
}

func (k *Kuwo) GetMediaUrl(meta miaosic.MetaData, quality miaosic.Quality) ([]miaosic.MediaUrl, error) {
	var lastErr error
	for _, providerQuality := range k.mediaURLQualities(quality) {
		mediaUrl, err := k.getMediaUrlConvertURL2(meta.Identifier, providerQuality)
		if err == nil {
			return []miaosic.MediaUrl{mediaUrl}, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (k *Kuwo) mediaURLQualities(quality miaosic.Quality) []miaosic.Quality {
	providerQuality := k.MapQuality(quality)
	if providerQuality == QualityFlac {
		return []miaosic.Quality{QualityFlac, QualityMp3}
	}
	return []miaosic.Quality{providerQuality}
}

func (k *Kuwo) getMediaUrlConvertURL2(id string, quality miaosic.Quality) (miaosic.MediaUrl, error) {
	format := "mp3"
	if quality == QualityFlac {
		format = "flac"
	}

	query := fmt.Sprintf("user=0&corp=kuwo&source=%s&p2p=1&type=convert_url2&sig=0&format=%s&rid=%s", kuwoConvertURL2Source, format, id)
	values := url.Values{}
	values.Set("f", "kuwo")
	values.Set("q", base64.StdEncoding.EncodeToString(Encrypt([]byte(query))))

	body, err := kuwoGetMobiBody("/mobi.s?" + values.Encode())
	if err != nil {
		return miaosic.MediaUrl{}, err
	}
	rawURL := kuwoDownloadURLRegex.FindString(string(body))
	if rawURL == "" {
		return miaosic.MediaUrl{}, miaosic.ErrorExternalApi
	}
	return miaosic.NewMediaUrl(rawURL, kuwoMiaosicQuality(quality)), nil
}

func kuwoGetMobiBody(path string) ([]byte, error) {
	conn, err := (&net.Dialer{Timeout: 5 * time.Second}).Dial("tcp", "mobi.kuwo.cn:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(8 * time.Second)); err != nil {
		return nil, err
	}
	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: mobi.kuwo.cn\r\nUser-Agent: okhttp/3.10.0\r\nConnection: close\r\n\r\n", path)
	if _, err := conn.Write([]byte(req)); err != nil {
		return nil, err
	}
	body, readErr := io.ReadAll(conn)
	if len(body) == 0 {
		if readErr != nil {
			return nil, readErr
		}
		return nil, miaosic.ErrorExternalApi
	}
	return body, nil
}

func kuwoMiaosicQuality(quality miaosic.Quality) miaosic.Quality {
	switch quality {
	case QualityMp3, Quality320kMp3:
		return miaosic.Quality320k
	case Quality256kMp3:
		return miaosic.Quality256k
	case Quality192kMp3:
		return miaosic.Quality192k
	case Quality128kMp3:
		return miaosic.Quality128k
	case QualityFlac:
		return miaosic.QualitySQ
	default:
		return miaosic.QualityStandard
	}
}

func (k *Kuwo) GetMediaLyric(meta miaosic.MetaData) ([]miaosic.Lyrics, error) {
	resp, err := k.client.R().
		SetHeaders(k.header).
		SetQueryParam("musicId", meta.Identifier).
		Get("http://m.kuwo.cn/newh5/singles/songinfoandlrc")
	if err != nil {
		return nil, err
	}
	respResult := gjson.ParseBytes(resp.Body())
	//pp.Println(resp.String())
	lrcs := make([]string, 0)
	respResult.Get("data.lrclist").ForEach(func(key, value gjson.Result) bool {
		lrcs = append(lrcs, fmt.Sprintf("[00:%s]%s", value.Get("time").String(), value.Get("lineLyric").String()))
		return true
	})
	if len(lrcs) == 0 {
		return nil, miaosic.ErrorExternalApi
	}
	return []miaosic.Lyrics{utils.ParseLyricWithLangDetection(strings.Join(lrcs, "\n"))}, nil
}

func (k *Kuwo) MatchMedia(keyword string) (miaosic.MetaData, bool) {
	if id := k.IdRegex0.FindString(keyword); id != "" {
		return miaosic.MetaData{
			Provider:   k.GetName(),
			Identifier: id,
		}, true
	}
	if id := k.IdRegex1.FindString(keyword); id != "" {
		return miaosic.MetaData{
			Provider:   k.GetName(),
			Identifier: id[2:],
		}, true
	}
	return miaosic.MetaData{}, false
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
	//d = 28886583
	n += strconv.Itoa(d)

	var num1, num2 int64
	for len(n) > 10 {
		// stupid javascript
		if len(n[10:]) > 19 {
			num1 = cast.ToInt64(n[10:11])
			// 19+8+1
			num2 = cast.ToInt64(n[19 : 19+8+1])
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
