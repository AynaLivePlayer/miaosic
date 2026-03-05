package kugou

import (
	"os"
	"strings"
	"testing"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/require"
)

var testApi = NewKugou(false)

func init() {
	_ = testApi.RestoreSession(os.Getenv("kugou_session"))
}

func TestKugou_Search(t *testing.T) {
	result, err := testApi.Search("心似烟火", 1, 20)
	require.NoError(t, err, "Search Error")
	require.NotEmpty(t, result, "Search Result Empty")
	require.Equal(t, 20, len(result), "Search Result Length")
	t.Log(result[0])
}

func TestKugou_MatchMedia(t *testing.T) {
	meta, ok := testApi.MatchMedia("24aae0ef48311770043044ab2376a8db")
	require.True(t, ok)
	require.Equal(t, "24aae0ef48311770043044ab2376a8db", meta.Identifier)
	meta, ok = testApi.MatchMedia("24aae0ef48311770043044ab2376a8dbaa")
	require.False(t, ok)
}

func TestKugou_GetMediaInfo(t *testing.T) {
	meta := miaosic.MetaData{Identifier: "c79c062ff4b362ac253031c6e577e722", Provider: testApi.GetName()}
	result, err := testApi.GetMediaInfo(meta)
	require.NoError(t, err, "GetMediaInfo Error")
	require.NotEmpty(t, result, "GetMediaInfo Result Empty")
	t.Log(result)
}

func TestKugou_GetMediaInfo2(t *testing.T) {
	meta := miaosic.MetaData{Identifier: "24aae0ef48311770043044ab2376a8db", Provider: testApi.GetName()}
	result, err := testApi.GetMediaInfo(meta)
	require.NoError(t, err, "GetMediaInfo Error")
	require.NotEmpty(t, result, "GetMediaInfo Result Empty")
	t.Log(result)
}

func TestKugou_GetMediaInfo3_Verify_Artists(t *testing.T) {
	// 雪
	meta := miaosic.MetaData{Identifier: "dd789e543aadbaeacf0c92aee592a601", Provider: testApi.GetName()}
	result, err := testApi.GetMediaInfo(meta)
	require.NoError(t, err, "GetMediaInfo Error")
	require.NotNil(t, result, "GetMediaInfo Result Empty")
	require.Equal(t, []string{"杜婧荧", "王艺翔"}, result.Artists)
	require.NotEmpty(t, result, "GetMediaInfo Result Empty")
	t.Log(result)
}

func TestKugou_GetMediaUrl(t *testing.T) {
	meta := miaosic.MetaData{Identifier: strings.ToLower("729723fa4afc5b33561f7f78489cacdf"), Provider: testApi.GetName()}
	result, err := testApi.GetMediaUrl(meta, miaosic.QualityAny)
	require.NoError(t, err, "GetMediaUrl Error")
	require.NotEmpty(t, result, "GetMediaUrl Result Empty")
	t.Log(result)
}

func TestKugou_GetMediaUrl_Vip(t *testing.T) {
	meta := miaosic.MetaData{Identifier: strings.ToLower("24aae0ef48311770043044ab2376a8db"), Provider: testApi.GetName()}
	result, err := testApi.GetMediaUrl(meta, miaosic.QualitySQ)
	require.NoError(t, err, "GetMediaUrl Error")
	require.NotEmpty(t, result, "GetMediaUrl Result Empty")
	t.Log(result)
}

func TestKugou_GetMediaUrlAcappella(t *testing.T) {
	meta := miaosic.MetaData{Identifier: strings.ToLower("16bbfe8fbcd9e62731c3c44ba79ae794"), Provider: testApi.GetName()}
	result, err := testApi.GetMediaUrl(meta, "magic_acappella")
	require.NoError(t, err, "GetMediaUrl Error")
	require.NotEmpty(t, result, "GetMediaUrl Result Empty")
	t.Log(result)
}

func TestKugou_GetMediaLyric(t *testing.T) {
	meta := miaosic.MetaData{Identifier: "24aae0ef48311770043044ab2376a8db", Provider: testApi.GetName()}
	result, err := testApi.GetMediaLyric(meta)
	require.NoError(t, err, "GetMediaLyric Error")
	require.NotEmpty(t, result, "GetMediaLyric Result Empty")
	t.Log(result)
}
