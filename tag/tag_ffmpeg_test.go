package tag

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadTagsWrittenByFFmpeg(t *testing.T) {
	requireFFmpeg(t)
	source, coverPath, cover := testDataPaths(t)

	for _, tc := range tagTestFormats() {
		t.Run(tc.name, func(t *testing.T) {
			meta := testMetadata(tc.name, cover)
			if !tc.ffmpegWritesCover {
				meta.Pictures = nil
			}
			path := filepath.Join(t.TempDir(), "ffmpeg-written"+tc.ext)
			writeTaggedSampleWithFFmpeg(t, tc, source, coverPath, path, meta)

			got := readMetadata(t, path)
			assertMetadata(t, meta, got, tc.ffprobeLyricsOK, tc.ffmpegWritesCover)
		})
	}
}

func TestWriteTagsVerifiedByFFmpeg(t *testing.T) {
	requireFFmpeg(t)
	source, _, cover := testDataPaths(t)

	for _, tc := range tagTestFormats() {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "sample"+tc.ext)
			writeCleanSampleWithFFmpeg(t, tc, source, path)
			meta := testMetadata(tc.name, cover)

			require.NoError(t, WriteTo(path, meta))
			assertFFProbeMetadata(t, path, meta, tc.ffprobeLyricsOK)
		})
	}
}

func writeTaggedSampleWithFFmpeg(t *testing.T, tc tagFormatCase, source, coverPath, path string, meta Metadata) {
	t.Helper()
	if tc.name == "ogg" {
		metaPath := filepath.Join(t.TempDir(), "metadata.ffmeta")
		require.NoError(t, os.WriteFile(metaPath, ffmetadata(meta), 0o600))
		args := []string{"-hide_banner", "-loglevel", "error", "-y", "-i", source, "-i", metaPath, "-map", "0:a:0", "-map_metadata", "1", "-t", "1", "-c:a", "libvorbis", path}
		run(t, "ffmpeg", args...)
		return
	}

	args := []string{"-hide_banner", "-loglevel", "error", "-y", "-i", source}
	if tc.ffmpegWritesCover {
		args = append(args, "-i", coverPath)
	}
	args = append(args, "-map", "0:a:0")
	if tc.ffmpegWritesCover {
		args = append(args, "-map", "1:v:0")
	}
	args = append(args, "-map_metadata", "-1", "-t", "1")
	args = append(args, codecArgsForFFmpegWrite(tc)...)
	if tc.ffmpegWritesCover {
		args = append(args, "-c:v", "copy", "-disposition:v:0", "attached_pic")
	}
	args = append(args,
		"-metadata", "title="+meta.Title,
		"-metadata", "artist="+meta.Artist,
		"-metadata", "album="+meta.Album,
		"-metadata", "lyrics="+meta.Lyrics[0].Lyrics,
		path,
	)
	run(t, "ffmpeg", args...)
}

func ffmetadata(meta Metadata) []byte {
	var buf strings.Builder
	buf.WriteString(";FFMETADATA1\n")
	writeFFMetadataLine(&buf, "title", meta.Title)
	writeFFMetadataLine(&buf, "artist", meta.Artist)
	writeFFMetadataLine(&buf, "album", meta.Album)
	writeFFMetadataLine(&buf, "lyrics", meta.Lyrics[0].Lyrics)
	if len(meta.Pictures) > 0 {
		writeFFMetadataLine(&buf, "METADATA_BLOCK_PICTURE", vorbisPictureBlock(meta.Pictures[0]))
	}
	return []byte(buf.String())
}

func writeFFMetadataLine(buf *strings.Builder, key, value string) {
	buf.WriteString(key)
	buf.WriteByte('=')
	buf.WriteString(strings.NewReplacer("\\", "\\\\", "\n", "\\\n", "=", "\\=", ";", "\\;", "#", "\\#").Replace(value))
	buf.WriteByte('\n')
}

func assertFFProbeMetadata(t *testing.T, path string, meta Metadata, checkLyrics bool) {
	t.Helper()
	out, err := exec.Command("ffprobe",
		"-hide_banner",
		"-loglevel", "error",
		"-show_entries", "format_tags:stream=codec_type:stream_tags:stream_disposition=attached_pic",
		"-of", "json",
		path,
	).CombinedOutput()
	require.NoErrorf(t, err, "ffprobe %s\n%s", path, out)

	var probe struct {
		Streams []struct {
			CodecType   string            `json:"codec_type"`
			Tags        map[string]string `json:"tags"`
			Disposition map[string]int    `json:"disposition"`
		} `json:"streams"`
		Format struct {
			Tags map[string]string `json:"tags"`
		} `json:"format"`
	}
	require.NoError(t, json.Unmarshal(out, &probe))

	tags := map[string]string{}
	addTags := func(src map[string]string) {
		for k, v := range src {
			key := strings.ToLower(k)
			if _, ok := tags[key]; !ok {
				tags[key] = v
			}
		}
	}
	addTags(probe.Format.Tags)
	hasCover := false
	for _, stream := range probe.Streams {
		addTags(stream.Tags)
		if stream.CodecType == "video" || stream.Disposition["attached_pic"] == 1 {
			hasCover = true
		}
	}
	if tags["metadata_block_picture"] != "" {
		hasCover = true
	}

	require.Equal(t, meta.Title, tags["title"])
	require.Equal(t, meta.Artist, tags["artist"])
	require.Equal(t, meta.Album, tags["album"])
	if checkLyrics {
		require.Equal(t, meta.Lyrics[0].Lyrics, tags["lyrics"])
	}
	require.True(t, hasCover, fmt.Sprintf("ffprobe should see embedded cover artwork in %s", path))
}
