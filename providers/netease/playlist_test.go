package netease

import (
	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/assert"
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
