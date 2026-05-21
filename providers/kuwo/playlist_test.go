package kuwo

import (
	"testing"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/require"
)

func TestKuwo_MatchPlaylist(t *testing.T) {
	meta, ok := api.MatchPlaylist("http://www.kuwo.cn/playlist_detail/2959147566")
	require.True(t, ok)
	require.Equal(t, api.GetName(), meta.Provider)
	require.Equal(t, "2959147566", meta.Identifier)

	meta, ok = api.MatchPlaylist("https://www.kuwo.cn/playlist/2959147566")
	require.True(t, ok)
	require.Equal(t, "2959147566", meta.Identifier)

	meta, ok = api.MatchPlaylist("2959147566")
	require.True(t, ok)
	require.Equal(t, "2959147566", meta.Identifier)

	_, ok = api.MatchPlaylist("https://www.kuwo.cn/play_detail/22804772")
	require.False(t, ok)
}

func TestKuwo_GetPlaylist(t *testing.T) {
	playlist, err := api.GetPlaylist(miaosic.MetaData{
		Provider:   api.GetName(),
		Identifier: "2959147566",
	})
	require.NoError(t, err)
	require.NotEmpty(t, playlist.Medias)
	require.Equal(t, api.GetName(), playlist.Meta.Provider)
	require.Equal(t, "2959147566", playlist.Meta.Identifier)
	t.Logf("playlist=%q medias=%d sample=%s/%s", playlist.Title, len(playlist.Medias), playlist.Medias[0].Title, playlist.Medias[0].Artist)
}
