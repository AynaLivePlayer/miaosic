package kugou

import (
	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/require"
	"testing"
)

var api miaosic.MediaProvider = NewKugou()

func TestKugou_Search(t *testing.T) {
	result, err := api.Search("心似烟火", 1, 20)
	require.NoError(t, err, "Search Error")
	require.NotEmpty(t, result, "Search Result Empty")
	require.Equal(t, 20, len(result), "Search Result Length")
	t.Log(result[0])
}

func TestKugou_GetMediaInfo(t *testing.T) {
	meta := miaosic.MetaData{Identifier: "c79c062ff4b362ac253031c6e577e722", Provider: api.GetName()}
	result, err := api.GetMediaInfo(meta)
	require.NoError(t, err, "GetMediaInfo Error")
	require.NotEmpty(t, result, "GetMediaInfo Result Empty")
	t.Log(result)
}

func TestKugou_GetMediaUrl(t *testing.T) {
	meta := miaosic.MetaData{Identifier: "c79c062ff4b362ac253031c6e577e722", Provider: api.GetName()}
	result, err := api.GetMediaUrl(meta, miaosic.QualityAny)
	require.NoError(t, err, "GetMediaUrl Error")
	require.NotEmpty(t, result, "GetMediaUrl Result Empty")
	t.Log(result)
}

func TestKugou_GetMediaLyric(t *testing.T) {
	meta := miaosic.MetaData{Identifier: "c79c062ff4b362ac253031c6e577e722", Provider: api.GetName()}
	result, err := api.GetMediaLyric(meta)
	require.NoError(t, err, "GetMediaLyric Error")
	require.NotEmpty(t, result, "GetMediaLyric Result Empty")
	t.Log(result)
}
