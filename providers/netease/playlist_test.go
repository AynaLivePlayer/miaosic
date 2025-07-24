package netease

import (
	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetPlaylist(t *testing.T) {
	playlist, err := api.GetPlaylist(miaosic.MetaData{
		Provider:   api.GetName(),
		Identifier: "2520739691",
		//Identifier: "2382819181",
	})
	assert.NoError(t, err)
	assert.True(t, len(playlist.Medias) >= 150)
	for _, media := range playlist.Medias {
		t.Log(media.Title, media.Artist, media.Album)
	}
}

func TestNetease_MatchPlaylist(t *testing.T) {
	meta, ok := api.MatchPlaylist("https://music.163.com/#/playlist?id=3204776824")
	assert.True(t, ok)
	require.Equal(t, api.GetName(), meta.Provider)
	require.Equal(t, "3204776824", meta.Identifier)
}
