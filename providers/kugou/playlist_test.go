package kugou

import (
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/require"
	"testing"
)

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
