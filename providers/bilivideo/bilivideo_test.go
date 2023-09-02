package bilivideo

import (
	"fmt"
	"github.com/aynakeya/deepcolor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"miaosic"
	"regexp"
	"strings"
	"testing"
)

var api miaosic.MediaProvider = NewBilibiliViedo(deepcolor.NewRestyRequester())

func TestBV_GetMusicMeta(t *testing.T) {
	media := api.MatchMedia("BV1434y1q71P")
	require.NotNil(t, media)
	require.NoError(t, api.UpdateMedia(media))
	assert.Equal(t, "卦者那啥子靈風", media.Artist)
}

func TestBV_GetMusic(t *testing.T) {
	media := api.MatchMedia("BV1434y1q71P")
	require.NoError(t, api.UpdateMedia(media))
	require.NoError(t, api.UpdateMediaUrl(media))
	assert.True(t, strings.Contains(media.Url, "bilivideo"), media.Url)
}

func TestBV_Regex(t *testing.T) {
	assert.Equal(t, "BV1gA411P7ir?p=3", regexp.MustCompile("^BV[0-9A-Za-z]+(\\?p=[0-9]+)?").FindString("BV1gA411P7ir?p=3"))
}

func TestBV_GetMusicMeta2(t *testing.T) {
	media := api.MatchMedia("BV1gA411P7ir?p=3")
	require.NotNil(t, media)
	require.NoError(t, api.UpdateMedia(media))
	require.Equal(t, "沈默沈默", media.Artist)
}

func TestBV_GetMusic2(t *testing.T) {
	media := api.MatchMedia("BV1gA411P7ir?p=3")
	require.NoError(t, api.UpdateMedia(media))
	require.NoError(t, api.UpdateMediaUrl(media))
	assert.Equal(t, "沈默沈默", media.Artist)
	fmt.Println(media.Url)
}

func TestBV_Search(t *testing.T) {
	result, err := api.Search("家有女友op")
	require.NoError(t, err, "Search Error")
	require.NotEmpty(t, result, "Search Result Empty")
	t.Log(result[0])
}
