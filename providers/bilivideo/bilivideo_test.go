package bilivideo

import (
	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"strings"
	"testing"
)

var api miaosic.MediaProvider = NewBilibiliViedo()

func TestBV_GetMusicMeta(t *testing.T) {
	meta, ok := api.MatchMedia("BV1434y1q71P")
	require.True(t, ok)
	media, err := api.GetMediaInfo(meta)
	require.NoError(t, err)
	require.Equal(t, "卦者那啥子靈風", media.Artist)
}

func TestBV_GetMusic(t *testing.T) {
	meta, _ := api.MatchMedia("BV1434y1q71P")
	urls, err := api.GetMediaUrl(meta, miaosic.QualityAny)
	require.NoError(t, err)
	require.NotEmpty(t, urls)
	assert.True(t, strings.Contains(urls[0].Url, "http"))
}

func TestBV_Regex(t *testing.T) {
	assert.Equal(t, "BV1gA411P7ir?p=3", regexp.MustCompile("^BV[0-9A-Za-z]+(\\?p=[0-9]+)?").FindString("BV1gA411P7ir?p=3"))
}

func TestBV_GetMusicMeta2(t *testing.T) {
	meta, ok := api.MatchMedia("BV1Ax4y1Y7et")
	require.True(t, ok)
	media, err := api.GetMediaInfo(meta)
	require.NoError(t, err)
	require.Equal(t, "言力华", media.Artist)
}

func TestBV_GetMusic2(t *testing.T) {
	meta, _ := api.MatchMedia("BV1gA411P7ir?p=3")
	urls, err := api.GetMediaUrl(meta, miaosic.QualityAny)
	require.NoError(t, err)
	require.NotEmpty(t, urls)
	assert.True(t, strings.Contains(urls[0].Url, "http"))
}

func TestBV_Search(t *testing.T) {
	result, err := api.Search("家有女友op", 1, 20)
	require.NoError(t, err, "Search Error")
	require.NotEmpty(t, result, "Search Result Empty")
	require.Equal(t, 20, len(result), "Search Result Length")
	t.Log(result[0])
}
