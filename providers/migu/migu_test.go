package migu

import (
	"fmt"
	"testing"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/require"
)

var testApi = NewMigu()

func formatMediaInfoForLog(info miaosic.MediaInfo) string {
	return fmt.Sprintf("title=%q artist=%q album=%q provider=%q identifier=%q", info.Title, info.Artist, info.Album, info.Meta.Provider, info.Meta.Identifier)
}

func formatMediaURLForLog(url miaosic.MediaUrl) string {
	return fmt.Sprintf("quality=%q url=%q headers=%d", url.Quality, url.Url, len(url.Header))
}

func formatLyricsForLog(lyrics miaosic.Lyrics) string {
	firstLine := ""
	if len(lyrics.Content) > 0 {
		firstLine = lyrics.Content[0].Lyric
	}
	return fmt.Sprintf("lang=%q lines=%d first=%q", lyrics.Lang, len(lyrics.Content), firstLine)
}

func TestMigu_Qualities(t *testing.T) {
	require.Equal(t, []miaosic.Quality{
		QualityLow,
		QualityStandard,
		QualityHigh,
		QualityLossless,
	}, testApi.Qualities())
	require.Equal(t, QualityHigh, testApi.MapQuality(miaosic.QualityAny))
	require.Equal(t, QualityStandard, testApi.MapQuality(miaosic.Quality128k))
	require.Equal(t, QualityHigh, testApi.MapQuality(miaosic.Quality320k))
	require.Equal(t, QualityLossless, testApi.MapQuality(miaosic.QualitySQ))
	require.Equal(t, QualityHigh, testApi.MapQuality("hq"))
}

func TestMigu_MatchMedia(t *testing.T) {
	meta, ok := testApi.MatchMedia("https://music.migu.cn/v3/music/song/600902000006889498")
	require.True(t, ok)
	require.Equal(t, testApi.GetName(), meta.Provider)
	require.Equal(t, "600902000006889498", meta.Identifier)

	meta, ok = testApi.MatchMedia("mg600902000006889498")
	require.True(t, ok)
	require.Equal(t, "600902000006889498", meta.Identifier)

	_, ok = testApi.MatchMedia("https://example.com/song/60054701934")
	require.False(t, ok)
}

func TestMigu_Search(t *testing.T) {
	result, err := testApi.Search("周杰伦", 1, 10)
	require.NoError(t, err, "Search Error")
	require.NotEmpty(t, result, "Search Result Empty")
	require.Len(t, result, 10)

	media := result[0]
	require.Equal(t, testApi.GetName(), media.Meta.Provider)
	require.NotEmpty(t, media.Meta.Identifier)
	require.NotEmpty(t, media.Title)
	require.NotEmpty(t, media.Artist)
	t.Log(formatMediaInfoForLog(media))
}

func TestMigu_SearchPagination(t *testing.T) {
	page1, err := testApi.Search("周杰伦", 1, 10)
	require.NoError(t, err, "Search Page1 Error")
	require.Len(t, page1, 10)

	page2, err := testApi.Search("周杰伦", 2, 10)
	require.NoError(t, err, "Search Page2 Error")
	require.Len(t, page2, 10)

	require.NotEqual(t, page1[0].Meta.Identifier, page2[0].Meta.Identifier)
	t.Log(fmt.Sprintf("page1=%s page2=%s", formatMediaInfoForLog(page1[0]), formatMediaInfoForLog(page2[0])))
}

func TestMigu_SearchThenGetMediaInfo(t *testing.T) {
	result, err := testApi.Search("周杰伦 晴天", 1, 10)
	require.NoError(t, err, "Search Error")
	require.NotEmpty(t, result, "Search Result Empty")

	media, err := testApi.GetMediaInfo(result[0].Meta)
	require.NoError(t, err, "GetMediaInfo Error")
	require.NotEmpty(t, media.Title)
	require.NotEmpty(t, media.Artist)
	require.Equal(t, testApi.GetName(), media.Meta.Provider)
	require.NotEmpty(t, media.Meta.Identifier)
	t.Log(formatMediaInfoForLog(media))
}

func TestMigu_GetMediaUrl(t *testing.T) {
	result, err := testApi.Search("周杰伦 晴天", 1, 10)
	require.NoError(t, err, "Search Error")
	require.NotEmpty(t, result, "Search Result Empty")

	_, err = testApi.GetMediaUrl(result[0].Meta, miaosic.QualityAny)
	require.Error(t, err, "ref music-lib listenSong endpoint currently does not return a playable url")
	t.Log(err)
}

func TestMigu_GetMediaUrl_WithMatchedMeta(t *testing.T) {
	meta, ok := testApi.MatchMedia("https://music.migu.cn/v3/music/song/600902000006889498")
	require.True(t, ok)

	_, err := testApi.GetMediaUrl(meta, miaosic.QualityAny)
	require.Error(t, err, "ref music-lib listenSong endpoint currently does not return a playable url")
	t.Log(err)
}

func TestMigu_GetMediaLyric(t *testing.T) {
	result, err := testApi.Search("周杰伦 晴天", 1, 10)
	require.NoError(t, err, "Search Error")
	require.NotEmpty(t, result, "Search Result Empty")

	lyrics, err := testApi.GetMediaLyric(result[0].Meta)
	require.NoError(t, err, "GetMediaLyric Error")
	require.NotEmpty(t, lyrics, "GetMediaLyric Result Empty")
	require.NotEmpty(t, lyrics[0].Content, "GetMediaLyric Content Empty")
	t.Log(formatLyricsForLog(lyrics[0]))
}
