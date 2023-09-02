package bilivideo

import (
	"AynaLivePlayer/core/adapter"
	"AynaLivePlayer/core/model"
	"github.com/stretchr/testify/assert"
	"regexp"
	"strings"
	"testing"
)

func TestBV_GetMusicMeta(t *testing.T) {
	var api adapter.MediaProvider = BilibiliVideoAPI

	media := model.Media{
		Meta: model.Meta{
			Name: api.GetName(),
			Id:   "BV1434y1q71P",
		},
	}
	err := api.UpdateMedia(&media)
	assert.Nil(t, err)
	assert.Equal(t, "卦者那啥子靈風", media.Artist)
}

func TestBV_GetMusic(t *testing.T) {
	var api adapter.MediaProvider = BilibiliVideoAPI
	media := model.Media{
		Meta: model.Meta{
			Name: api.GetName(),
			Id:   "BV1434y1q71P",
		},
	}
	err := api.UpdateMedia(&media)
	assert.Nil(t, err)
	err = api.UpdateMediaUrl(&media)
	assert.Nil(t, err)
	assert.True(t, strings.Contains(media.Url, "bilivideo"), media.Url)
}

func TestBV_Regex(t *testing.T) {
	assert.Equal(t, "BV1gA411P7ir?p=3", regexp.MustCompile("^BV[0-9A-Za-z]+(\\?p=[0-9]+)?").FindString("BV1gA411P7ir?p=3"))
}

func TestBV_GetMusicMeta2(t *testing.T) {
	var api adapter.MediaProvider = BilibiliVideoAPI

	media := model.Media{
		Meta: model.Meta{
			Name: api.GetName(),
			Id:   "BV1gA411P7ir?p=3",
		},
	}
	err := api.UpdateMedia(&media)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, "沈默沈默", media.Artist)
}

func TestBV_GetMusic2(t *testing.T) {
	var api adapter.MediaProvider = BilibiliVideoAPI
	media := model.Media{
		Meta: model.Meta{
			Name: api.GetName(),
			Id:   "BV1gA411P7ir?p=1",
		},
	}
	err := api.UpdateMedia(&media)
	assert.Nil(t, err)
	err = api.UpdateMediaUrl(&media)
	assert.Nil(t, err)
	assert.Equal(t, "沈默沈默", media.Artist)
}

func TestBV_Search(t *testing.T) {
	var api adapter.MediaProvider = BilibiliVideoAPI
	result, err := api.Search("家有女友")
	assert.Nil(t, err, "Search Error")
	assert.Truef(t, len(result) > 0, "Search Result Empty")
	t.Log(result[0])
}
