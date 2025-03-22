package kugou

import (
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKugou_MatchPlaylist(t *testing.T) {
	meta, ok := testApi.MatchPlaylist("https://m.kugou.com/share/zlist.html?listid=2&type=0&uid=600319512&share_type=collect&from=pcCode&_t=795992922&global_collection_id=collection_3_600319512_2_0&sign=b10567180f66e08d562f5142a8f1f8b9&chain=5JZSIebEnV3")
	require.True(t, ok)
	require.Equal(t, "collection_3_600319512_2_0", meta.Identifier)

}

func TestKugou_GetPlaylist(t *testing.T) {
	// less than 100 song
	playlist, err := testApi.GetPlaylist(miaosic.MetaData{Identifier: "gcid_3zfcfgjcz31z06d"})
	require.NoError(t, err)
	fmt.Println(playlist.Medias)
}

func TestKugou_GetPlaylist_2(t *testing.T) {
	// more than 100 song
	playlist, err := testApi.GetPlaylist(miaosic.MetaData{Identifier: "gcid_3ztimg53zoz09e"})
	require.NoError(t, err)
	fmt.Println(playlist.Medias)
}

func TestKugou_GetPlaylist_3(t *testing.T) {
	playlist, err := testApi.GetPlaylist(miaosic.MetaData{Identifier: "collection_3_600319512_2_0"})
	require.NoError(t, err)
	fmt.Println(playlist.Medias)
	fmt.Println(playlist.Title)
}

func TestKugou_getCollectionId(t *testing.T) {
	val, err := testApi.getCollectionId("gcid_3zfcfgjcz31z06d")
	require.NoError(t, err)
	require.Equal(t, "collection_3_806499027_106_0", val)
}

func TestKugou_getCollectionId_2(t *testing.T) {
	val, err := testApi.getCollectionId("gcid_3ztimg53zoz09e")
	require.NoError(t, err)
	require.Equal(t, "collection_3_1551108653_24_0", val)
}

func TestKugou_getPlaylistTitle(t *testing.T) {
	val, err := testApi.getPlaylistTitle("collection_3_806499027_106_0")
	require.NoError(t, err)
	require.Equal(t, "emo伤感天花板｜来自0.8×的孤独与失恋", val)
}
