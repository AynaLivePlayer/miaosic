package netease

import (
	"os"
	"strings"
	"testing"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/require"
)

var api = NewNetease()

func init() {
	os.Setenv("netease_session", "eyJNVVNJQ19VIjoiMDA1RUVFNEQ5NzI1N0E2M0YwRkY0MzYzMEVDNThENTZGRkRGQTFBQ0IzOUQyREJENDIxNThEQTg1NjFGNTg2NDlCOTIxNDZDREIwMkZCMTU2Qzc4QjMwOThCQTNGOEExMzdGNkU4NEVDQzM2Q0RDMEU0MzlCOTk1NEU3OEVBNTNFMEEyOTdCOUEwM0I4OUM4MURDODdBNEUwMkI0M0Y4QUUyODFCQUM4Q0QwQzNDRTdFRERFQTgzQkU0MEY5Q0U5QzREODUwNTlBNzNDMjgwQjY0NjQyODU4MDQzNkJCRUQ1REM1NzFDREQ5MUIzODcwQzU5NkRGQjg3ODJGRkEyMDg1Q0JGODM3MDRCOTU3NjdGRTI4QkNGMjdGODM3MkIzRTBCNkM1NzY5MzgzOUZEMzQxNjFCMTMyNkQ0NTVGNDNDQTgyNTMxQ0Q4QzcwMTJCNjg1RThFM0E1RTJDMDFFMjNGRDY4QzAyODk3MkVCMUU1MUNBRjMyMzJCOEU5MkNGRkMyMzc5Nzg1RERFMDNENzQxMDMzQzJGRTZBNzE0RUYxRDk2MkI1MkVGN0EyRTYyODVEM0NCQ0QyMkE2OTk2Q0NFN0MwREU0ODgwM0YxQjM1M0NBNzQ2MzcwODJBODlENDFFNDA1QTEyNDMyNzNGMEZGMkUzQ0VGODA0NTgxQTFDMzZDRDE4Mjk5ODdDNkJFNjUzNkZEREYxQjIwMjA1RjVEMUI5NTVFQUIzREM2NTYyNzlFNjMyOTZCNjY3NjBFMzNCRERBNkNGN0YzOUU0NjA3MTE2N0UxOTk3RjY3QUUwNjZFM0U2M0IxMTc5NkY4MDBCRUY0QTIyRDg3RDM2MzVCOTkwOEJDQTYyMzFEQkFCMkMxNzMwOEVBREU3NERBNEQ1QTc5RDc5MTE5NDU2Q0FGNzczMTAxNEQ3NzUyRkJFQjBDREU2RUM4QkY1MUQ0ODQxOEJEMUUyRThGREQzRDBCOTgwQTYzMzZCMjc4NDA3RTZFMUQ5N0EyNjYyNUVENjE0QzFGOUZGQzA4NTQ5QUYxNDdDMDExM0I4M0I3NTkzMjQwM0RGNDdFRTYxNjQ4NjQ3MjhFOTcyMkY4QzcyQ0RBQkZDODQzM0VFMUVFMDI5MzFCNEQ1ODI0MzRBMEFFNUUiLCJfX2NzcmYiOiI0YWM0NTBlYjVmZjRhNzAxNWRlYjljNDVlZjU2NTUwMCIsImRldmljZUlkIjoicmR3c2JzN213dnBqcWgwaXhzaXhrNTlpbzFnZnZ6bGQifQ==")
	api.RestoreSession(os.Getenv("netease_session"))
}

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
	require.Equal(t, []string{"花粥", "王胜娚"}, result[0].Artists)
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

func TestNetease_GetMusicMetaRadioProgram(t *testing.T) {
	//meta := miaosic.MetaData{
	//	Provider:   api.GetName(),
	//	Identifier: "3321110387",
	//}
	// todo finish radio program
	//media, err := api.GetMediaInfo(meta)
	//require.NoError(t, err)
	//require.Equal(t, "染", media.Title)
	//require.Equal(t, "Reol", media.Artist)
	//t.Log(media.Cover.Url)
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

func TestNetease_GetMediaUrl_Login(t *testing.T) {
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

// todo: unable to get url for this song
// 无版权歌曲
func TestNetease_GetMediaUrl_Vip(t *testing.T) {
	meta := miaosic.MetaData{
		Provider:   api.GetName(),
		Identifier: "28038055",
	}
	_, err := api.GetMediaUrl(meta, miaosic.QualityAny)
	require.Error(t, err)
	//require.NotEmpty(t, urls)
	//require.True(t, strings.Contains(urls[0].Url, "http"))
	//t.Log(urls[0].Url)
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
