package qq

import (
	"github.com/AynaLivePlayer/miaosic"
	"github.com/k0kubun/pp/v3"
	"github.com/stretchr/testify/require"
	"testing"
)

var testApi *QQMusicProvider

func init() {
	testApi = NewQQMusicProvider()
}

func TestQQ_MatchMedia(t *testing.T) {
	result, ok := testApi.MatchMedia("002pCkT73uKyPL")
	require.True(t, ok)
	require.Equal(t, "002pCkT73uKyPL", result.Identifier)
	require.Equal(t, testApi.GetName(), result.Provider)
}

func TestQQ_Search(t *testing.T) {
	result, err := testApi.Search("还是会想你 h3R3", 1, 10)
	require.NoError(t, err, "Search Error")
	require.NotEmpty(t, result, "Search Result Empty")
	require.Equal(t, 10, len(result), "Search Result Length")
	//pp.Println(result)
}

func TestQQ_GetMediaInfo(t *testing.T) {
	meta := miaosic.MetaData{Identifier: "002pCkT73uKyPL", Provider: testApi.GetName()}
	result, err := testApi.GetMediaInfo(meta)
	require.NoError(t, err, "GetMediaInfo Error")
	require.NotEmpty(t, result, "GetMediaInfo Result Empty")
	require.Equal(t, "还是会想你", result.Title)
	pp.Println(result)
}

func TestQQ_GetMediaUrl(t *testing.T) {
	meta := miaosic.MetaData{Identifier: "002pCkT73uKyPL", Provider: testApi.GetName()}
	result, err := testApi.GetMediaUrl(meta, QualityOGG192)
	require.NoError(t, err, "GetMediaUrl Error")
	require.NotEmpty(t, result, "GetMediaUrl Result Empty")
	t.Log(result)
}
