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

	provider := NewLocal(localDir)
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

func TestLocalRealDataTestfolder(t *testing.T) {
	localDir := findLocalTestDataDir(t)
	start := time.Now()
	provider := NewLocal(localDir)
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
