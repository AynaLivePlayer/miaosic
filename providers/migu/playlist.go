package migu

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/AynaLivePlayer/miaosic"
)

func (m *Migu) MatchPlaylist(uri string) (miaosic.MetaData, bool) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`playlistId=(\d+)`),
		regexp.MustCompile(`musicListId=(\d+)`),
		regexp.MustCompile(`(?:playlist|songlist)/(\d+)`),
	}
	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(uri)
		if len(matches) >= 2 {
			return miaosic.NewMetaData(m.GetName(), matches[1]), true
		}
	}
	if strings.TrimSpace(uri) != "" && !strings.Contains(uri, "/") {
		return miaosic.NewMetaData(m.GetName(), strings.TrimSpace(uri)), true
	}
	return miaosic.MetaData{}, false
}

func (m *Migu) GetPlaylist(meta miaosic.MetaData) (*miaosic.Playlist, error) {
	if meta.Provider != "" && meta.Provider != m.GetName() {
		return nil, miaosic.ErrorInvalidMediaMeta
	}
	playlistID := strings.TrimSpace(meta.Identifier)
	if playlistID == "" {
		return nil, miaosic.ErrorInvalidMediaMeta
	}

	playlist := &miaosic.Playlist{
		Meta:   meta,
		Title:  "Migu Playlist " + playlistID,
		Medias: make([]miaosic.MediaInfo, 0),
	}
	if title, err := m.fetchPlaylistTitle(playlistID); err == nil && title != "" {
		playlist.Title = title
	}

	const pageSize = 50
	seen := make(map[string]struct{})
	totalCount := 0
	for pageNo := 1; ; pageNo++ {
		resp, err := m.client.R().
			SetHeaders(m.requestHeaders()).
			SetQueryParams(map[string]string{
				"pageNo":     strconv.Itoa(pageNo),
				"pageSize":   strconv.Itoa(pageSize),
				"playlistId": playlistID,
			}).
			Get("https://app.c.nf.migu.cn/MIGUM3.0/resource/playlist/song/v2.0")
		if err != nil {
			return nil, err
		}

		var result struct {
			Code string `json:"code"`
			Info string `json:"info"`
			Data struct {
				SongList   []miguSongItem `json:"songList"`
				TotalCount int            `json:"totalCount"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body(), &result); err != nil {
			return nil, fmt.Errorf("migu: playlist json parse error: %w", err)
		}
		if result.Code != "" && result.Code != "000000" {
			return nil, fmt.Errorf("migu: playlist api error %s (code %s)", result.Info, result.Code)
		}
		if totalCount == 0 {
			totalCount = result.Data.TotalCount
		}
		if len(result.Data.SongList) == 0 {
			break
		}

		before := len(playlist.Medias)
		for _, item := range result.Data.SongList {
			media, ok := m.convertItemToMedia(item, miaosic.QualityAny, true)
			if !ok {
				continue
			}
			key, _, _ := splitMiguIdentifier(media.Meta.Identifier)
			if key == "" {
				key = media.Title + "|" + media.Artist
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			playlist.Medias = append(playlist.Medias, media)
		}

		if len(result.Data.SongList) < pageSize {
			break
		}
		if totalCount > 0 && len(playlist.Medias) >= totalCount {
			break
		}
		if len(playlist.Medias) == before {
			break
		}
	}

	if len(playlist.Medias) == 0 {
		return nil, errors.New("migu: playlist has no playable songs")
	}
	return playlist, nil
}

func (m *Migu) fetchPlaylistTitle(playlistID string) (string, error) {
	resp, err := m.client.R().
		SetHeaders(m.requestHeaders()).
		SetQueryParams(map[string]string{
			"needSimple":   "00",
			"resourceType": "2021",
			"resourceId":   playlistID,
		}).
		Get("https://app.c.nf.migu.cn/MIGUM2.0/v1.0/content/resourceinfo.do")
	if err != nil {
		return "", err
	}

	var result struct {
		Code     string `json:"code"`
		Info     string `json:"info"`
		Resource []struct {
			Title string `json:"title"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", fmt.Errorf("migu: playlist info json parse error: %w", err)
	}
	if result.Code != "" && result.Code != "000000" {
		return "", fmt.Errorf("migu: playlist info api error %s (code %s)", result.Info, result.Code)
	}
	if len(result.Resource) == 0 {
		return "", errors.New("migu: playlist info not found")
	}
	return strings.TrimSpace(result.Resource[0].Title), nil
}
