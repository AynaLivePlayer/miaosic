package netease

import (
	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

var api = NewNetease()

func TestNetease_Search(t *testing.T) {
	result, err := api.Search("染 reol", 1, 20)
	require.NoError(t, err)
	require.NotEmpty(t, result)
	t.Log(result[0].Title, result[0].Artist, result[0].Album)
}

func TestNetease_Search2(t *testing.T) {
	result, err := api.Search("出山", 1, 20)
	require.NoError(t, err)
	media := result[0]
	require.Equal(t, "花粥,王胜娚", result[0].Artist)
	urls, err := api.GetMediaUrl(media.Meta, miaosic.QualityAny)
	require.NoError(t, err)
	require.NotEmpty(t, urls)
	require.True(t, strings.Contains(urls[0].Url, "http"))
	t.Log(urls[0].Url)
}

func TestNetease_GetMusicMeta(t *testing.T) {
	meta := miaosic.MetaData{
		Provider:   api.GetName(),
		Identifier: "33516503",
	}
	media, err := api.GetMediaInfo(meta)
	require.NoError(t, err)
	require.Equal(t, "染", media.Title)
	require.Equal(t, "Reol", media.Artist)
	t.Log(media.Cover.Url)
}

func TestNetease_UpdateMediaLyric(t *testing.T) {
	meta := miaosic.MetaData{
		Provider:   api.GetName(),
		Identifier: "33516503",
	}
	lrcs, err := api.GetMediaLyric(meta)
	require.NoError(t, err)
	require.NotEmpty(t, lrcs)
}

// todo: unable to get url for this song
func TestNetease_GetMediaUrl(t *testing.T) {
	meta := miaosic.MetaData{
		Provider:   api.GetName(),
		Identifier: "2601642780",
	}
	urls, err := api.GetMediaUrl(meta, miaosic.QualityAny)
	require.NoError(t, err)
	require.NotEmpty(t, urls)
	require.True(t, strings.Contains(urls[0].Url, "http"))
	t.Log(urls[0].Url)
}

//func TestNetease_GetPlaylist(t *testing.T) {
//	var api adapter.MediaProvider = NeteaseAPI
//	playlist, err := api.GetPlaylist(&model.Meta{
//		Name: api.GetName(),
//		//Id:   "2520739691",
//		Id: "2382819181",
//	})
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	fmt.Println(len(playlist))
//	for _, media := range playlist {
//		fmt.Println(media.Title, media.Artist, media.Album)
//	}
//
//}
