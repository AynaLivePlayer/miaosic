package tag

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type tagFormatCase struct {
	name              string
	ext               string
	args              []string
	ffprobeLyricsOK   bool
	ffmpegWritesCover bool
}

func tagTestFormats() []tagFormatCase {
	return []tagFormatCase{
		{name: "mp3", ext: ".mp3", args: []string{"-map", "0:a:0", "-map_metadata", "-1", "-t", "1", "-c:a", "libmp3lame"}, ffprobeLyricsOK: true, ffmpegWritesCover: true},
		{name: "m4a", ext: ".m4a", args: []string{"-map", "0:a:0", "-map_metadata", "-1", "-t", "1", "-c:a", "aac"}, ffprobeLyricsOK: true, ffmpegWritesCover: true},
		{name: "aac", ext: ".aac", args: []string{"-map", "0:a:0", "-map_metadata", "-1", "-t", "1", "-c:a", "aac", "-f", "adts"}, ffprobeLyricsOK: true},
		{name: "ogg", ext: ".ogg", args: []string{"-map", "0:a:0", "-map_metadata", "-1", "-t", "1", "-c:a", "libvorbis"}, ffprobeLyricsOK: true, ffmpegWritesCover: true},
		{name: "flac", ext: ".flac", args: []string{"-map", "0:a:0", "-map_metadata", "-1", "-t", "1", "-c:a", "flac"}, ffprobeLyricsOK: true, ffmpegWritesCover: true},
		{name: "wav", ext: ".wav", args: []string{"-map", "0:a:0", "-map_metadata", "-1", "-t", "1", "-c:a", "pcm_s16le"}},
	}
}

func testDataPaths(t *testing.T) (string, string, []byte) {
	t.Helper()
	root := filepath.Clean("..")
	source := filepath.Join(root, "data", "origin.wav")
	coverPath := filepath.Join(root, "data", "cover.jpg")
	cover, err := os.ReadFile(coverPath)
	require.NoError(t, err)
	return source, coverPath, cover
}

func testMetadata(name string, cover []byte) Metadata {
	return Metadata{
		Title:  "miaosic title " + name,
		Artist: "miaosic artist",
		Album:  "miaosic album",
		Lyrics: []Lyrics{{Lang: "eng", Lyrics: "roundtrip lyrics " + name}},
		Pictures: []Picture{{
			Mimetype:    "image/jpeg",
			Type:        PictureTypeFrontCover,
			Description: "front cover",
			Data:        cover,
		}},
	}
}

func writeCleanSampleWithFFmpeg(t *testing.T, tc tagFormatCase, source, path string) {
	t.Helper()
	ffmpegArgs := append([]string{"-hide_banner", "-loglevel", "error", "-y", "-i", source}, tc.args...)
	ffmpegArgs = append(ffmpegArgs, path)
	run(t, "ffmpeg", ffmpegArgs...)
}

func codecArgsForFFmpegWrite(tc tagFormatCase) []string {
	switch tc.name {
	case "mp3":
		return []string{"-c:a", "libmp3lame"}
	case "m4a":
		return []string{"-c:a", "aac"}
	case "aac":
		return []string{"-c:a", "aac", "-f", "adts", "-write_id3v2", "1"}
	case "flac":
		return []string{"-c:a", "flac"}
	case "wav":
		return []string{"-c:a", "pcm_s16le"}
	default:
		return tc.args
	}
}

func readMetadata(t *testing.T, path string) Metadata {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	got, err := Read(f)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return got
}

func assertMetadata(t *testing.T, expected Metadata, got Metadata, checkLyrics, checkCover bool) {
	t.Helper()
	require.Equal(t, expected.Title, got.Title)
	require.Equal(t, expected.Artist, got.Artist)
	require.Equal(t, expected.Album, got.Album)
	if checkLyrics {
		require.NotEmpty(t, got.Lyrics)
		require.Equal(t, expected.Lyrics[0].Lyrics, got.Lyrics[0].Lyrics)
	}
	if checkCover {
		require.NotEmpty(t, got.Pictures)
		require.Equal(t, "image/jpeg", got.Pictures[0].Mimetype)
		require.True(t, bytes.Equal(expected.Pictures[0].Data, got.Pictures[0].Data), "cover image should round trip")
	}
}

func requireFFmpeg(t *testing.T) {
	t.Helper()
	_, err := exec.LookPath("ffmpeg")
	require.NoError(t, err)
	_, err = exec.LookPath("ffprobe")
	require.NoError(t, err)
}

func run(t *testing.T, name string, args ...string) {
	t.Helper()
	out, err := exec.Command(name, args...).CombinedOutput()
	require.NoErrorf(t, err, "%s %v\n%s", name, args, out)
}
