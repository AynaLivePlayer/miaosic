package kuwo

import (
	"fmt"
	"github.com/aynakeya/deepcolor"
	"github.com/stretchr/testify/require"
	"miaosic"
	"testing"
)

var api miaosic.MediaProvider = NewKuwo(deepcolor.NewRestyRequester())

func TestKuwo_Secret(t *testing.T) {
	// using 80378195 as d
	require.Equal(t, "5add7ba59bc95d3d38a8983a82af6efc78d3484c6d253f29a2154dd042b3383604ca7953",
		api.(*Kuwo).generateSecret("zddnb2yWCXJjk6aWb2tSZBNeaPBChEPY", "Hm_Iuvt_cdb524f42f0cer9b268e4v7y734w5esq24"))

}

func TestKuwo_Search(t *testing.T) {
	result, err := api.Search("周杰伦")
	require.NoError(t, err)
	fmt.Println(result)
	media := result[0]
	err = api.UpdateMediaUrl(media)
	fmt.Println(err)
	fmt.Println(media.Url)
}

func TestKuwo_GetMusicMeta(t *testing.T) {
	media := miaosic.Media{
		Meta: miaosic.MediaMeta{
			Provider:   api.GetName(),
			Identifier: "22804772",
		},
	}
	err := api.UpdateMedia(&media)
	require.NoError(t, err)
	require.Equal(t, "霜雪千年", media.Title)
}

func TestKuwo_GetMusic(t *testing.T) {
	media := miaosic.Media{
		Meta: miaosic.MediaMeta{
			Provider:   api.GetName(),
			Identifier: "22804772",
		},
	}
	require.NoError(t, api.UpdateMedia(&media))
	require.NoError(t, api.UpdateMediaUrl(&media))
	require.Equal(t, "霜雪千年", media.Title)
	require.True(t, len(media.Url) > 0)
}

//func TestKuwo_UpdateMediaLyric(t *testing.T) {
//	var api adapter.MediaProvider = providers.KuwoAPI
//	media := model.Media{
//		Meta: model.Meta{
//			Name: api.GetName(),
//			Id:   "22804772",
//		},
//	}
//	err := api.UpdateMediaLyric(&media)
//	fmt.Println(err)
//	fmt.Println(media.Lyric)
//}
//
//func TestKuwo_GetPlaylist(t *testing.T) {
//	var api adapter.MediaProvider = providers.KuwoAPI
//	playlist, err := api.GetPlaylist(&model.Meta{
//		Name: api.GetName(),
//		//Id:   "1082685104",
//		Id: "2959147566",
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
