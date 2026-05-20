package local

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/require"
)

func TestLocalScansWithoutReadingTags(t *testing.T) {
	localDir := t.TempDir()
	playlistName := "demo"
	playlistDir := filepath.Join(localDir, playlistName)
	require.NoError(t, os.MkdirAll(playlistDir, 0755))

	for i := 0; i < 128; i++ {
		name := fmt.Sprintf("song-%03d.mp3", i)
		require.NoError(t, os.WriteFile(filepath.Join(playlistDir, name), []byte("not an mp3"), 0644))
	}
	require.NoError(t, os.WriteFile(filepath.Join(playlistDir, "song-ignored.lrc"), []byte("[00:00]ignored"), 0644))

	provider := NewLocal(localDir, DontScanTag, DontUseLocalCache)
	playlist, err := provider.GetPlaylist(miaosic.MetaData{
		Provider:   provider.GetName(),
		Identifier: playlistName,
	})
	require.NoError(t, err)
	require.Len(t, playlist.Medias, 128)

	result, err := provider.Search("song-042", 1, 10)
	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Equal(t, path.Join(playlistName, "song-042.mp3"), result[0].Meta.Identifier)

	info, err := provider.GetMediaInfo(result[0].Meta)
	require.NoError(t, err)
	require.Equal(t, "song-042.mp3", info.Title)
	require.Equal(t, "Unknown", info.Artist)
}

func TestNewLocalDefaultOptions(t *testing.T) {
	provider := NewLocal(t.TempDir())
	require.Equal(t, DontScanTag, provider.scanMode)
	require.Equal(t, DontUseLocalCache, provider.cacheMode)
}

func TestLocalGetPlaylistRefreshesDirectory(t *testing.T) {
	localDir := t.TempDir()
	playlistName := "demo"
	playlistDir := filepath.Join(localDir, playlistName)
	require.NoError(t, os.MkdirAll(playlistDir, 0755))
	firstPath := filepath.Join(playlistDir, "first.mp3")
	secondPath := filepath.Join(playlistDir, "second.mp3")
	require.NoError(t, os.WriteFile(firstPath, []byte("not an mp3"), 0644))

	provider := NewLocal(localDir, DontScanTag, DontUseLocalCache)
	playlist, err := provider.GetPlaylist(miaosic.MetaData{Provider: provider.GetName(), Identifier: playlistName})
	require.NoError(t, err)
	require.Len(t, playlist.Medias, 1)

	require.NoError(t, os.WriteFile(secondPath, []byte("not an mp3"), 0644))
	playlist, err = provider.GetPlaylist(miaosic.MetaData{Provider: provider.GetName(), Identifier: playlistName})
	require.NoError(t, err)
	require.Len(t, playlist.Medias, 2)

	require.NoError(t, os.Remove(firstPath))
	playlist, err = provider.GetPlaylist(miaosic.MetaData{Provider: provider.GetName(), Identifier: playlistName})
	require.NoError(t, err)
	require.Len(t, playlist.Medias, 1)
	require.Equal(t, path.Join(playlistName, "second.mp3"), playlist.Medias[0].Meta.Identifier)
}

func TestLocalSearchPaginatesFromIndex(t *testing.T) {
	provider := &Local{
		playlists: map[string]*localPlaylist{
			"indexed": {
				name: "indexed",
				medias: []localMedia{
					newIndexedTestMedia("indexed/song-001.mp3", "song alpha"),
					newIndexedTestMedia("indexed/song-002.mp3", "song beta"),
					newIndexedTestMedia("indexed/song-003.mp3", "song gamma"),
				},
			},
		},
		mediaByID: make(map[string]*localMedia),
	}
	provider.rebuildIndexes()

	result, err := provider.Search("", 2, 1)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "indexed/song-002.mp3", result[0].Meta.Identifier)
}

func TestLocalTagScanModes(t *testing.T) {
	localDir := t.TempDir()
	playlistName := "demo"
	playlistDir := filepath.Join(localDir, playlistName)
	require.NoError(t, os.MkdirAll(playlistDir, 0755))
	for i := 0; i < 4; i++ {
		require.NoError(t, os.WriteFile(filepath.Join(playlistDir, fmt.Sprintf("song-%d.mp3", i)), []byte("not an mp3"), 0644))
	}

	withoutTags := NewLocal(localDir, DontScanTag, DontUseLocalCache)
	require.False(t, allTagsAttempted(withoutTags, playlistName))

	withStartupTags := NewLocal(localDir, ScanTagOnStartup, DontUseLocalCache)
	require.True(t, allTagsAttempted(withStartupTags, playlistName))
	require.False(t, anyCoverAttempted(withStartupTags, playlistName))

	withBackgroundTags := NewLocal(localDir, ScanTagInBackground, DontUseLocalCache)
	require.Eventually(t, func() bool {
		return allTagsAttempted(withBackgroundTags, playlistName)
	}, time.Second, 10*time.Millisecond)
	require.False(t, anyCoverAttempted(withBackgroundTags, playlistName))
}

func TestLocalTagCache(t *testing.T) {
	localDir := t.TempDir()
	playlistName := "demo"
	filename := "cached.mp3"
	identifier := path.Join(playlistName, filename)
	playlistDir := filepath.Join(localDir, playlistName)
	require.NoError(t, os.MkdirAll(playlistDir, 0755))
	audioPath := filepath.Join(playlistDir, filename)
	require.NoError(t, os.WriteFile(audioPath, []byte("audio-data"), 0644))
	fileInfo, err := os.Stat(audioPath)
	require.NoError(t, err)

	require.NoError(t, writeLocalTagCache(filepath.Join(localDir, localCacheFileName), localTagCache{
		Version: 1,
		Medias: map[string]localTagCacheMedia{
			identifier: {
				Size:        fileInfo.Size(),
				ModTimeNano: fileInfo.ModTime().UnixNano(),
				Title:       "Cached Title",
				Artist:      "ArtistFromCacheOnly",
				Album:       "Cached Album",
			},
		},
	}))

	provider := NewLocal(localDir, DontScanTag, UseLocalCache)
	result, err := provider.Search("ArtistFromCacheOnly", 1, 10)
	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Equal(t, identifier, result[0].Meta.Identifier)
	require.Equal(t, "Cached Title", result[0].Title)
	require.True(t, allTagsAttempted(provider, playlistName))
	require.False(t, anyCoverAttempted(provider, playlistName))

	require.NoError(t, os.WriteFile(audioPath, []byte("changed-audio-data"), 0644))
	provider = NewLocal(localDir, DontScanTag, UseLocalCache)
	result, err = provider.Search("ArtistFromCacheOnly", 1, 10)
	require.NoError(t, err)
	require.Empty(t, result)
	require.False(t, allTagsAttempted(provider, playlistName))
}

func TestLocalTagCacheAppliesToNewPlaylist(t *testing.T) {
	localDir := t.TempDir()
	playlistName := "demo"
	filename := "late.mp3"
	identifier := path.Join(playlistName, filename)
	provider := NewLocal(localDir, DontScanTag, UseLocalCache)

	playlistDir := filepath.Join(localDir, playlistName)
	require.NoError(t, os.MkdirAll(playlistDir, 0755))
	audioPath := filepath.Join(playlistDir, filename)
	require.NoError(t, os.WriteFile(audioPath, []byte("audio-data"), 0644))
	fileInfo, err := os.Stat(audioPath)
	require.NoError(t, err)
	require.NoError(t, writeLocalTagCache(filepath.Join(localDir, localCacheFileName), localTagCache{
		Version: 1,
		Medias: map[string]localTagCacheMedia{
			identifier: {
				Size:        fileInfo.Size(),
				ModTimeNano: fileInfo.ModTime().UnixNano(),
				Title:       "Late Cached Title",
				Artist:      "LateCacheArtistOnly",
				Album:       "Late Cached Album",
			},
		},
	}))

	playlist, err := provider.GetPlaylist(miaosic.MetaData{Provider: provider.GetName(), Identifier: playlistName})
	require.NoError(t, err)
	require.Len(t, playlist.Medias, 1)
	require.Equal(t, "Late Cached Title", playlist.Medias[0].Title)
	result, err := provider.Search("LateCacheArtistOnly", 1, 10)
	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Equal(t, identifier, result[0].Meta.Identifier)
}

func TestLocalTagCacheSkipsFailedScans(t *testing.T) {
	localDir := t.TempDir()
	playlistName := "demo"
	playlistDir := filepath.Join(localDir, playlistName)
	require.NoError(t, os.MkdirAll(playlistDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(playlistDir, "song.mp3"), []byte("not an mp3"), 0644))

	provider := NewLocal(localDir, ScanTagOnStartup, UseLocalCache)
	cache, err := readLocalTagCache(provider.cachePath())
	require.NoError(t, err)
	require.NotContains(t, cache.Medias, path.Join(playlistName, "song.mp3"))
}

func TestLocalRealDataTestfolder(t *testing.T) {
	localDir := findLocalTestDataDir(t)
	start := time.Now()
	provider := NewLocal(localDir, DontScanTag, DontUseLocalCache)
	startupDuration := time.Since(start)

	start = time.Now()
	playlist, err := provider.GetPlaylist(miaosic.MetaData{
		Provider:   provider.GetName(),
		Identifier: "Vocaloid",
	})
	scanDuration := time.Since(start)
	require.NoError(t, err)
	require.Greater(t, len(playlist.Medias), 1000)

	start = time.Now()
	result, err := provider.Search("晴夜", 1, 10)
	searchDuration := time.Since(start)
	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Less(t, startupDuration, time.Second)
	require.Less(t, scanDuration, time.Second)
	require.Less(t, searchDuration, 100*time.Millisecond)
	t.Logf("playlist=%d startup=%s scan=%s search=%s sample=%s", len(playlist.Medias), startupDuration, scanDuration, searchDuration, result[0].Title)
}

func TestLocalRealDataReadTagsOnStartup(t *testing.T) {
	localDir := findLocalTestDataDir(t)
	start := time.Now()
	provider := NewLocal(localDir, ScanTagOnStartup, DontUseLocalCache)
	duration := time.Since(start)

	playlist, err := provider.GetPlaylist(miaosic.MetaData{
		Provider:   provider.GetName(),
		Identifier: "Vocaloid",
	})
	require.NoError(t, err)
	require.Greater(t, len(playlist.Medias), 1000)
	require.True(t, allTagsAttempted(provider, "Vocaloid"))
	require.False(t, anyCoverAttempted(provider, "Vocaloid"))
	t.Logf("read tags playlist=%d duration=%s sample=%s/%s", len(playlist.Medias), duration, playlist.Medias[0].Title, playlist.Medias[0].Artist)
}

func BenchmarkLocalSearchLargeLibrary(b *testing.B) {
	docs := make([]localSearchDoc, 0, 50000)
	for i := 0; i < 50000; i++ {
		info := miaosic.MediaInfo{
			Title:  fmt.Sprintf("song-%05d", i),
			Artist: fmt.Sprintf("artist-%03d", i%100),
			Album:  fmt.Sprintf("album-%02d", i%20),
			Meta: miaosic.MetaData{
				Provider:   "local",
				Identifier: fmt.Sprintf("bench/song-%05d.mp3", i),
			},
		}
		docs = append(docs, localSearchDoc{
			info:   info,
			search: localSearchText(info),
		})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := rankLocalSearchDocs("song-49999 artist-099", docs)
		if len(result) == 0 {
			b.Fatal("empty search result")
		}
	}
}

func allTagsAttempted(provider *Local, playlistName string) bool {
	provider.mu.RLock()
	defer provider.mu.RUnlock()
	playlist := provider.playlists[playlistName]
	if playlist == nil || len(playlist.medias) == 0 {
		return false
	}
	for idx := range playlist.medias {
		if !playlist.medias[idx].tagScanned {
			return false
		}
	}
	return true
}

func anyCoverAttempted(provider *Local, playlistName string) bool {
	provider.mu.RLock()
	defer provider.mu.RUnlock()
	playlist := provider.playlists[playlistName]
	if playlist == nil {
		return false
	}
	for idx := range playlist.medias {
		if playlist.medias[idx].coverLoaded {
			return true
		}
	}
	return false
}

func findLocalTestDataDir(t *testing.T) string {
	t.Helper()
	for _, dir := range []string{
		"data/testfolder",
		filepath.Join("..", "..", "data", "testfolder"),
	} {
		if stat, err := os.Stat(filepath.Join(dir, "Vocaloid")); err == nil && stat.IsDir() {
			return dir
		}
	}
	t.Skip("data/testfolder/Vocaloid not found")
	return ""
}

func newIndexedTestMedia(identifier, title string) localMedia {
	info := miaosic.MediaInfo{
		Title:  title,
		Artist: "artist",
		Album:  "album",
		Meta: miaosic.MetaData{
			Provider:   "local",
			Identifier: identifier,
		},
	}
	return localMedia{
		info:   info,
		search: localSearchText(info),
	}
}
