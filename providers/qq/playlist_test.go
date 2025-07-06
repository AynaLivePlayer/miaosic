package qq

import (
	"github.com/k0kubun/pp/v3"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQQMusicProvider_MatchPlaylist(t *testing.T) {
	val1, ok1 := testApi.MatchPlaylist("https://y.qq.com/n/ryqq/playlist/9515850830")
	require.True(t, ok1)
	require.Equal(t, "9515850830", val1.Identifier)
	require.Equal(t, testApi.GetName(), val1.Provider)
	val2, ok2 := testApi.MatchPlaylist("9515850830")
	require.True(t, ok2)
	require.Equal(t, "9515850830", val2.Identifier)
	require.Equal(t, testApi.GetName(), val2.Provider)
}

func TestQQMusicProvider_GetPlaylist(t *testing.T) {
	val1, ok1 := testApi.MatchPlaylist("https://y.qq.com/n/ryqq/playlist/7426999757")
	require.True(t, ok1)
	playlist, err := testApi.GetPlaylist(val1)
	require.NoError(t, err)
	require.True(t, len(playlist.Medias) >= 33)
	pp.Println(playlist)
}
