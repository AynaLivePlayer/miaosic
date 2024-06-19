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
		Identifier: "6536164",
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
	url := "http://mobi.kuwo.cn/mobi.s?f=kuwo&q=" + base64.StdEncoding.EncodeToString(Encrypt([]byte("source=jiakong&p2p=1&sig=1476474&type=convert_url_with_sign&format="+format+"&rid="+"6536164"+br)))
	result, err := resty.New().R().Get(url)
	fmt.Println(err)
	fmt.Println(result.String())
}

func TestKuwo_DesDecCar(t *testing.T) {
	val1, err := base64.StdEncoding.DecodeString("OCGk_ECm6dOb_DJVKzHfKPA1qKoKNG8sz1IhuP8pKvw1GoN4mZM6LYX9iFdgEkJfYXBrNAGAe2eDwZLE3Vxpj41B5q0s6z8RLqmUPkXZ2CRyBvo9VR6-IF6-4iGy9C_D3hIBq4xG2abDBxpjMepcJE_O2VVWGhdtXrNBuPjZqBGxggi2M-Xx8_fHHPe3gIXXfhJMle10Orc3B8KbuxSI2s8xhcwOqQJM2MAuK6VXJpl_jBrQZyi1sFxJVx4DRXGCntYe6V0I3RVG5V4Z3rUZBIMq0N9NMYDrVFIEE5RyBJfaPUS4SVRQzdY5aUxFEhnZHlR3QLgLvCwaO74tZQANy3VnigEuuBANVHcZm5bpWByEZgH_urF_q9mhOZEFfqirxvcNhQxNiI_Ng6Yr8lcZLZCecrz0lZ-yc2o3C1DmB_OIwFWSdxU6K6OcpBOnzOrdR6O74I_NPcDA7Jy1PsBQ5WHwbI73nMhsnw97300cte2WDbqzn2d45ZkHjftAP2MSmSxHUNacNp2zgHOfakTF8hbVxHfMzzNt6vOIl6TAMOAxs7awl9ZDLMPAEeKQhAykMr6aa1FhHmk2xNEh9Uvkx4VvHUCGvYxUr0uS2iYZq7KLoHvIWoOm9Cr8_S25OjwiNzTiayfCdIcFDOCkiPVIIT-eEtWfxFoQu-iwfb0DF31TijGSvGsxnGorZSbbd1XpnMm3tmyFRdfL1_15K7erWFh2GLu-_TASR1wXUoIkgLhSyMyk-x_xdT-DhXyTGZqWTTVs8Rl52QGDE62WTesahmn7Phl7MKnpKz0LgsD2Q0-zsd4c6AzQ8DQfN0YliqOa")
	require.NoError(t, err)
	val := Decrypt(val1)
	fmt.Println(string(val))
}

func TestKuwo_CarApi(t *testing.T) {
	//query := "user=C_APK_guanwang_e1699663d6664f1bbd04faad379c290c&randomcode=e1699663d6664f1bbd04faad379c290c&prod=kwplayercar_ar_6.3.9.40&corp=kuwo&vipver=8.5.5.0&source=kwplayercar_ar_6.3.9.40_C_APK_guanwang.apk&packageName=cn.kuwo.kwmusiccar&packageSign=c29a5a2ca235f326a4800335ecafbf31ace2d0a2&q36=f8cc39a5ae393f56713d909010001d418608&loginUid=&loginSid=&carModel=null&type=lyric&songname=%E7%AC%AC%E5%85%AD%E6%84%9F&artist=Reol&filename=&duration=191000&req=2&lrcx=1&rid=146301111&encode=utf8"
	//query := "vipver=8.5.5.0&source=kwplayercar_ar_6.3.9.40_C_APK_guanwang.apk&loginUid=&loginSid=&carModel=null&type=lyric&songname=%E7%AC%AC%E5%85%AD%E6%84%9F&artist=Reol&filename=&duration=191000&req=2&lrcx=1&rid=146301111&encode=utf8"
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
