package kuwo

import (
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/require"
	"testing"
)

var api miaosic.MediaProvider = NewKuwo()

//func TestKuwo_Secret(t *testing.T) {
//	// using 80378195 as d
//	require.Equal(t, "5add7ba59bc95d3d38a8983a82af6efc78d3484c6d253f29a2154dd042b3383604ca7953",
//		api.(*Kuwo).generateSecret("zddnb2yWCXJjk6aWb2tSZBNeaPBChEPY", "Hm_Iuvt_cdb524f42f0cer9b268e4v7y734w5esq24"))
//
//}

func TestKuwo_Search(t *testing.T) {
	result, err := api.Search("周杰伦", 1, 20)
	require.NoError(t, err)
	fmt.Println(result)
	media := result[0]
	urls, err := api.GetMediaUrl(media.Meta, miaosic.QualityAny)
	require.NoError(t, err)
	fmt.Println(urls)
}

func TestKuwo_GetMusicMeta(t *testing.T) {
	meta := miaosic.MediaMeta{
		Provider:   api.GetName(),
		Identifier: "22804772",
	}
	info, err := api.GetMediaInfo(meta)
	require.NoError(t, err)
	require.Equal(t, "霜雪千年", info.Title)
}

func TestKuwo_GetMusic(t *testing.T) {
	meta := miaosic.MediaMeta{
		Provider:   api.GetName(),
		Identifier: "22804772",
	}
	urls, err := api.GetMediaUrl(meta, miaosic.QualityAny)
	require.NoError(t, err)
	require.True(t, len(urls) > 0)
}

func TestKuwo_UpdateMediaLyric(t *testing.T) {
	meta := miaosic.MediaMeta{
		Provider:   api.GetName(),
		Identifier: "22804772",
	}
	lyrics, err := api.GetMediaLyric(meta)
	require.NoError(t, err)
	require.NotEmpty(t, len(lyrics) > 0)
}

//func TestKuwo_GetPlaylist(t *testing.T) {
//	playlist := miaosic.Playlist{
//		Meta: miaosic.MediaMeta{
//			Provider:   api.GetName(),
//			Identifier: "2959147566",
//		},
//	}
//	err := api.UpdatePlaylist(&playlist)
//	require.NoError(t, err)
//	require.NotEmpty(t, playlist.Medias)
//	t.Logf("sucessfully get %d medias", len(playlist.Medias))
//}
