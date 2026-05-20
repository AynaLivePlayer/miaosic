package migu

import (
	"fmt"
	"testing"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/require"
)

func TestMigu_MatchPlaylist(t *testing.T) {
	meta, ok := testApi.MatchPlaylist("https://music.migu.cn/v5/#/playlist?playlistId=233754996&playlistType=ordinary")
	require.True(t, ok)
	require.Equal(t, testApi.GetName(), meta.Provider)
	require.Equal(t, "233754996", meta.Identifier)

	meta, ok = testApi.MatchPlaylist("musicListId=233754996")
	require.True(t, ok)
	require.Equal(t, "233754996", meta.Identifier)

	_, ok = testApi.MatchPlaylist("https://example.com/playlist")
	require.False(t, ok)
}

func TestMigu_GetPlaylist(t *testing.T) {
	playlist, err := testApi.GetPlaylist(miaosic.MetaData{
		Provider:   testApi.GetName(),
		Identifier: "233754996",
	})
	require.NoError(t, err, "GetPlaylist Error")
	require.NotEmpty(t, playlist.Title, "Playlist Title Empty")
	require.NotEmpty(t, playlist.Medias, "Playlist Medias Empty")
	require.Greater(t, len(playlist.Medias), 0)
	require.LessOrEqual(t, len(playlist.Medias), 1000)
	t.Log(fmt.Sprintf("title=%q count=%d sample=%s", playlist.Title, len(playlist.Medias), formatMediaInfoForLog(playlist.Medias[0])))
}
