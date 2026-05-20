package bilibili

import (
	"testing"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/require"
)

var api miaosic.MediaProvider = NewBilibili()

func TestBilibili_Search(t *testing.T) {
	result, err := api.Search("染 reol", 1, 100)
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

func TestBilibili_GetMusic(t *testing.T) {
	meta := miaosic.MetaData{
		Provider:   api.GetName(),
		Identifier: "1560601",
	}
	_, err := api.GetMediaInfo(meta)
	require.NoError(t, err)
	urls, err := api.GetMediaUrl(meta, miaosic.QualityAny)
	require.NoError(t, err)
	require.NotEmpty(t, urls)
	require.NotEmpty(t, urls[0].Url)
	t.Log(urls[0].Url)
}
