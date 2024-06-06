package kuwo

import (
	"encoding/base64"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

var api miaosic.MediaProvider = NewKuwo()

//func TestKuwo_Secret(t *testing.T) {
//	// using 80378195 as d
//	require.Equal(t, "5add7ba59bc95d3d38a8983a82af6efc78d3484c6d253f29a2154dd042b3383604ca7953",
//		api.(*Kuwo).generateSecret("zddnb2yWCXJjk6aWb2tSZBNeaPBChEPY", "Hm_Iuvt_cdb524f42f0cer9b268e4v7y734w5esq24"))
//
//}

func TestKuwo_Secret2(t *testing.T) {
	// using 80378195 as d
	t.Log(
		api.(*Kuwo).generateSecret("Z3mc22m5FG2cezTznhS6YPNMPD5HnzSn", "Hm_Iuvt_cdb524f42f23cer9b268564v7y735ewrq2324"))

}

func TestKuwo_Search(t *testing.T) {
	result, err := api.Search("周杰伦", 1, 20)
	require.NoError(t, err)
	require.NotEmpty(t, result)
	media := result[0]
	require.True(t, strings.Contains(media.Artist, "周杰伦"))
	urls, err := api.GetMediaUrl(media.Meta, miaosic.QualityAny)
	require.NoError(t, err)
	require.NotEmpty(t, urls)
	require.True(t, strings.HasPrefix(urls[0].Url, "http"))
	t.Log(urls)
}

func TestKuwo_GetMusicMeta(t *testing.T) {
	meta := miaosic.MetaData{
		Provider:   api.GetName(),
		Identifier: "22804772",
	}
	info, err := api.GetMediaInfo(meta)
	require.NoError(t, err)
	require.Equal(t, "霜雪千年", info.Title)
}

func TestKuwo_GetMusic(t *testing.T) {
	meta := miaosic.MetaData{
		Provider:   api.GetName(),
		Identifier: "22804772",
	}
	urls, err := api.GetMediaUrl(meta, miaosic.QualityAny)
	require.NoError(t, err)
	require.NotEmpty(t, urls)
	t.Log(urls)
}

func TestKuwo_GetMusic2(t *testing.T) {
	meta := miaosic.MetaData{
		Provider:   api.GetName(),
		Identifier: "6510096",
	}
	urls, err := api.GetMediaUrl(meta, miaosic.QualityAny)
	require.NoError(t, err)
	require.NotEmpty(t, urls)

	t.Log(urls)
}

func TestKuwo_UpdateMediaLyric(t *testing.T) {
	meta := miaosic.MetaData{
		Provider:   api.GetName(),
		Identifier: "22804772",
	}
	lyrics, err := api.GetMediaLyric(meta)
	require.NoError(t, err)
	// Not sure
	require.NotEmpty(t, lyrics)
	require.NotEmpty(t, lyrics[0].Content)
}

// https://github.com/cnsilvan/UnblockNeteaseMusic/blob/master/provider/kuwo/kuwo.go
// http://anymatch.kuwo.cn/mobi.s?f=kwxs&q=
// http://mobi.kuwo.cn/mobi.s?f=kuwo&q=
func TestKuwo_Url(t *testing.T) {
	format := "mp3"
	br := "&br=128kmp3"
	url := "http://mobi.kuwo.cn/mobi.s?f=kuwo&q=" + base64.StdEncoding.EncodeToString(Encrypt([]byte("source=jiakong&p2p=1&sig=1476474&type=convert_url_with_sign&format="+format+"&rid="+"MP3_146301111"+br)))
	result, err := resty.New().R().Get(url)
	fmt.Println(err)
	fmt.Println(result.String())
}

//func TestKuwo_GetPlaylist(t *testing.T) {
//	playlist := miaosic.Playlist{
//		Meta: miaosic.MetaData{
//			Provider:   api.GetName(),
//			Identifier: "2959147566",
//		},
//	}
//	err := api.UpdatePlaylist(&playlist)
//	require.NoError(t, err)
//	require.NotEmpty(t, playlist.Medias)
//	t.Logf("sucessfully get %d medias", len(playlist.Medias))
//}
