package bilivideo

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBilibiliVideo_MatchPlaylist_Ok(t *testing.T) {
	meta, ok := api.MatchPlaylist("https://space.bilibili.com/346563107/favlist?fid=1179446107&ftype=create")
	require.True(t, ok)
	require.Equal(t, api.GetName(), meta.Provider)
	require.Equal(t, playlistFav+"_1179446107", meta.Identifier)
	meta, ok = api.MatchPlaylist("https://space.bilibili.com/346563107/channel/collectiondetail?sid=1889103&ctype=0")
	require.True(t, ok)
	require.Equal(t, api.GetName(), meta.Provider)
	require.Equal(t, playlistCollection+"_1889103", meta.Identifier)
}

func TestBilibiliVideo_MatchPlaylist_Fail(t *testing.T) {
	meta, ok := api.MatchPlaylist("https://space.bilibili.com/346563107")
	require.False(t, ok)
	require.Empty(t, meta)
	meta, ok = api.MatchPlaylist("https://space.bilibili.com/346563107/favlist")
	require.False(t, ok)
	require.Empty(t, meta)
	meta, ok = api.MatchPlaylist("https://space.bilibili.com/346563107/channel/collectiondetail")
	require.False(t, ok)
	require.Empty(t, meta)
}

func TestBilibiliVideo_GetPlaylist_Collection(t *testing.T) {
	uri := "https://space.bilibili.com/346563107/channel/collectiondetail?sid=1889103&ctype=0"
	meta, ok := api.MatchPlaylist(uri)
	require.True(t, ok)
	playlist, err := api.GetPlaylist(meta)
	require.NoError(t, err)
	require.Equal(t, meta, playlist.Meta)
	require.Equal(t, "合集·巴以冲突-世界在关注什么？", playlist.Title)
	require.GreaterOrEqual(t, len(playlist.Medias), 71)
}

func TestBilibiliVideo_GetPlaylist_Fav(t *testing.T) {
	uri := "https://space.bilibili.com/10003632/favlist?fid=729246932&ftype=create"
	meta, ok := api.MatchPlaylist(uri)
	require.True(t, ok)
	playlist, err := api.GetPlaylist(meta)
	require.NoError(t, err)
	require.Equal(t, meta, playlist.Meta)
	require.Equal(t, "AMV", playlist.Title)
	require.GreaterOrEqual(t, len(playlist.Medias), 12)
}

func TestBilibiliVideo_GetPlaylist_Fav2(t *testing.T) {
	uri := "https://space.bilibili.com/691501918/favlist?fid=3159910518&ftype=create"
	meta, ok := api.MatchPlaylist(uri)
	require.True(t, ok)
	playlist, err := api.GetPlaylist(meta)
	require.NoError(t, err)
	require.Equal(t, meta, playlist.Meta)
	require.Equal(t, "歌", playlist.Title)
	require.GreaterOrEqual(t, len(playlist.Medias), 26)
}

func TestBilibiliVideo_GetPlaylist_Fav3(t *testing.T) {
	uri := "https://space.bilibili.com/10003632/favlist?fid=56376932&ftype=create"
	meta, ok := api.MatchPlaylist(uri)
	require.True(t, ok)
	playlist, err := api.GetPlaylist(meta)
	require.NoError(t, err)
	require.Equal(t, meta, playlist.Meta)
	require.Equal(t, "Vocaloid Music", playlist.Title)
	require.Equal(t, len(playlist.Medias), 881)
}
