package kuwo

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/AynaLivePlayer/miaosic"
)

var (
	kuwoPlaylistIDRegexes = []*regexp.Regexp{
		regexp.MustCompile(`playlist_detail/(\d+)`),
		regexp.MustCompile(`playlist/(\d+)`),
		regexp.MustCompile(`(?:^|[?&])pid=(\d+)`),
	}
	kuwoPlaylistRawIDRegex = regexp.MustCompile(`^\d+$`)
)

func (k *Kuwo) MatchPlaylist(uri string) (miaosic.MetaData, bool) {
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return miaosic.MetaData{}, false
	}
	if kuwoPlaylistRawIDRegex.MatchString(uri) {
		return miaosic.NewMetaData(k.GetName(), uri), true
	}
	for _, pattern := range kuwoPlaylistIDRegexes {
		matches := pattern.FindStringSubmatch(uri)
		if len(matches) >= 2 {
			return miaosic.NewMetaData(k.GetName(), matches[1]), true
		}
	}
	return miaosic.MetaData{}, false
}

func (k *Kuwo) GetPlaylist(meta miaosic.MetaData) (*miaosic.Playlist, error) {
	if meta.Provider != "" && meta.Provider != k.GetName() {
		return nil, miaosic.ErrorInvalidMediaMeta
	}
	playlistID := strings.TrimSpace(meta.Identifier)
	if playlistID == "" {
		return nil, miaosic.ErrorInvalidMediaMeta
	}

	return k.getPlaylistFromNPLServer(playlistID)
}

func (k *Kuwo) getPlaylistFromNPLServer(playlistID string) (*miaosic.Playlist, error) {
	params := url.Values{}
	params.Set("op", "getlistinfo")
	params.Set("pid", playlistID)
	params.Set("pn", "0")
	params.Set("rn", "100")
	params.Set("encode", "utf8")
	params.Set("keyset", "pl2012")
	params.Set("identity", "kuwo")
	params.Set("pcmp4", "1")
	params.Set("vipver", "1")
	params.Set("newver", "1")

	resp, err := k.client.R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36").
		SetHeader("Cookie", k.header["cookie"]).
		Get("http://nplserver.kuwo.cn/pl.svc?" + params.Encode())
	if err != nil {
		return nil, err
	}

	playlist, err := k.parseNPLServerPlaylist(playlistID, resp.Body())
	if err != nil {
		return nil, err
	}
	return playlist, nil
}

func (k *Kuwo) parseNPLServerPlaylist(playlistID string, body []byte) (*miaosic.Playlist, error) {
	var result struct {
		Name      string `json:"name"`
		Title     string `json:"title"`
		ListName  string `json:"listname"`
		Info      string `json:"info"`
		Desc      string `json:"desc"`
		Img       string `json:"img"`
		Pic       string `json:"pic"`
		MusicList []struct {
			ID         string `json:"id"`
			MusicRID   string `json:"musicrid"`
			Name       string `json:"name"`
			SongName   string `json:"song_name"`
			Artist     string `json:"artist"`
			ArtistName string `json:"artist_name"`
			Album      string `json:"album"`
			AlbumPic   string `json:"albumpic"`
			Pic        string `json:"pic"`
			Pic120     string `json:"pic120"`
		} `json:"musiclist"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("kuwo: playlist detail json parse error: %w", err)
	}
	if len(result.MusicList) == 0 {
		return nil, errors.New("kuwo: playlist is empty or id is invalid")
	}

	playlist := &miaosic.Playlist{
		Title:  kuwoFirstNonEmpty(result.Name, result.Title, result.ListName, "Kuwo Playlist "+playlistID),
		Meta:   miaosic.NewMetaData(k.GetName(), playlistID),
		Medias: make([]miaosic.MediaInfo, 0, len(result.MusicList)),
	}
	seen := make(map[string]struct{}, len(result.MusicList))
	for _, item := range result.MusicList {
		rid := strings.TrimPrefix(kuwoFirstNonEmpty(item.ID, item.MusicRID), "MUSIC_")
		if rid == "" {
			continue
		}
		if _, ok := seen[rid]; ok {
			continue
		}
		seen[rid] = struct{}{}
		playlist.Medias = append(playlist.Medias, miaosic.MediaInfo{
			Title:  kuwoNormalizeText(kuwoFirstNonEmpty(item.Name, item.SongName)),
			Artist: kuwoNormalizeText(kuwoFirstNonEmpty(item.Artist, item.ArtistName)),
			Album:  kuwoNormalizeText(item.Album),
			Cover: miaosic.Picture{
				Url: kuwoNormalizeImageURL(kuwoFirstNonEmpty(item.AlbumPic, item.Pic120, item.Pic)),
			},
			Meta: miaosic.NewMetaData(k.GetName(), rid),
		})
	}
	if len(playlist.Medias) == 0 {
		return nil, errors.New("kuwo: playlist has no valid songs")
	}
	return playlist, nil
}

func kuwoFirstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func kuwoNormalizeText(value string) string {
	value = html.UnescapeString(value)
	value = strings.ReplaceAll(value, "\u00a0", " ")
	return strings.TrimSpace(value)
}

func kuwoNormalizeImageURL(raw string) string {
	raw = kuwoNormalizeText(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "//") {
		raw = "http:" + raw
	} else if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		switch {
		case strings.HasPrefix(raw, "img"):
			raw = "http://" + raw
		default:
			raw = "http://img1.kuwo.cn/star/albumcover/" + strings.TrimPrefix(raw, "/")
		}
	}
	replacements := []struct {
		old string
		new string
	}{
		{"/120/", "/500/"},
		{"/150/", "/500/"},
		{"/240/", "/500/"},
		{"_100.", "_500."},
		{"_120.", "_500."},
		{"_150.", "_500."},
		{"_240.", "_500."},
	}
	for _, replacement := range replacements {
		if strings.Contains(raw, replacement.old) {
			raw = strings.Replace(raw, replacement.old, replacement.new, 1)
		}
	}
	return raw
}
