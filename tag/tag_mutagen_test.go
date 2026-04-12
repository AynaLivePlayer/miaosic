package tag

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadTagsWrittenByMutagen(t *testing.T) {
	requireFFmpeg(t)
	requireMutagen(t)
	source, coverPath, cover := testDataPaths(t)

	for _, tc := range tagTestFormats() {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "mutagen-written"+tc.ext)
			writeCleanSampleWithFFmpeg(t, tc, source, path)
			meta := testMetadata(tc.name, cover)

			writeTaggedSampleWithMutagen(t, tc, path, coverPath, meta)
			got := readMetadata(t, path)
			assertMetadata(t, meta, got, true, true)
		})
	}
}

func TestWriteTagsVerifiedByMutagen(t *testing.T) {
	requireFFmpeg(t)
	requireMutagen(t)
	source, _, cover := testDataPaths(t)

	for _, tc := range tagTestFormats() {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "sample"+tc.ext)
			writeCleanSampleWithFFmpeg(t, tc, source, path)
			meta := testMetadata(tc.name, cover)

			require.NoError(t, WriteTo(path, meta))
			got := readTaggedSampleWithMutagen(t, tc, path)
			assertMutagenMetadata(t, meta, got)
		})
	}
}

type mutagenMetadata struct {
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	Lyrics    string `json:"lyrics"`
	CoverMIME string `json:"cover_mime"`
	CoverLen  int    `json:"cover_len"`
}

func writeTaggedSampleWithMutagen(t *testing.T, tc tagFormatCase, path, coverPath string, meta Metadata) {
	t.Helper()
	payload := map[string]string{
		"format":     tc.name,
		"path":       path,
		"cover_path": coverPath,
		"title":      meta.Title,
		"artist":     meta.Artist,
		"album":      meta.Album,
		"lyrics":     meta.Lyrics[0].Lyrics,
	}
	payloadPath := filepath.Join(t.TempDir(), "mutagen-write.json")
	writeJSON(t, payloadPath, payload)
	runPython(t, mutagenWriteScript, payloadPath)
}

func readTaggedSampleWithMutagen(t *testing.T, tc tagFormatCase, path string) mutagenMetadata {
	t.Helper()
	payload := map[string]string{
		"format": tc.name,
		"path":   path,
	}
	payloadPath := filepath.Join(t.TempDir(), "mutagen-read.json")
	writeJSON(t, payloadPath, payload)
	out := runPython(t, mutagenReadScript, payloadPath)
	var got mutagenMetadata
	require.NoError(t, json.Unmarshal(out, &got))
	return got
}

func assertMutagenMetadata(t *testing.T, expected Metadata, got mutagenMetadata) {
	t.Helper()
	require.Equal(t, expected.Title, got.Title)
	require.Equal(t, expected.Artist, got.Artist)
	require.Equal(t, expected.Album, got.Album)
	require.Equal(t, expected.Lyrics[0].Lyrics, got.Lyrics)
	require.Equal(t, "image/jpeg", got.CoverMIME)
	require.True(t, got.CoverLen > 0, "mutagen should see embedded cover artwork")
	require.Equal(t, len(expected.Pictures[0].Data), got.CoverLen)
}

func requireMutagen(t *testing.T) {
	t.Helper()
	_, err := exec.LookPath("python3")
	require.NoError(t, err)
	out, err := exec.Command("python3", "-c", "import mutagen; print(mutagen.version_string)").CombinedOutput()
	require.NoErrorf(t, err, "python mutagen is required: %s", out)
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o600))
}

func runPython(t *testing.T, script string, args ...string) []byte {
	t.Helper()
	cmdArgs := append([]string{"-c", script}, args...)
	out, err := exec.Command("python3", cmdArgs...).CombinedOutput()
	require.NoErrorf(t, err, "python3 mutagen script failed:\n%s", out)
	return bytes.TrimSpace(out)
}

const mutagenWriteScript = `
import base64
import json
import sys

from mutagen.flac import FLAC, Picture
from mutagen.id3 import APIC, ID3, TALB, TIT2, TPE1, TXXX, USLT
from mutagen.mp4 import MP4, MP4Cover
from mutagen.oggvorbis import OggVorbis
from mutagen.wave import WAVE

cfg = json.load(open(sys.argv[1], "r", encoding="utf-8"))
fmt = cfg["format"]
path = cfg["path"]
cover = open(cfg["cover_path"], "rb").read()
title = cfg["title"]
artist = cfg["artist"]
album = cfg["album"]
lyrics = cfg["lyrics"]

def add_id3_frames(tags):
    tags.add(TIT2(encoding=3, text=title))
    tags.add(TPE1(encoding=3, text=artist))
    tags.add(TALB(encoding=3, text=album))
    tags.add(USLT(encoding=3, lang="eng", desc="", text=lyrics))
    tags.add(TXXX(encoding=3, desc="lyrics", text=lyrics))
    tags.add(APIC(encoding=3, mime="image/jpeg", type=3, desc="front cover", data=cover))

def picture():
    pic = Picture()
    pic.type = 3
    pic.mime = "image/jpeg"
    pic.desc = "front cover"
    pic.data = cover
    return pic

if fmt in ("mp3", "aac"):
    tags = ID3()
    add_id3_frames(tags)
    tags.save(path, v2_version=4)
elif fmt == "wav":
    audio = WAVE(path)
    if audio.tags is None:
        audio.add_tags()
    audio.tags.clear()
    add_id3_frames(audio.tags)
    audio.save()
elif fmt == "m4a":
    audio = MP4(path)
    if audio.tags is None:
        audio.add_tags()
    audio.tags.clear()
    audio["\xa9nam"] = [title]
    audio["\xa9ART"] = [artist]
    audio["\xa9alb"] = [album]
    audio["\xa9lyr"] = [lyrics]
    audio["covr"] = [MP4Cover(cover, imageformat=MP4Cover.FORMAT_JPEG)]
    audio.save()
elif fmt == "flac":
    audio = FLAC(path)
    audio.clear()
    audio["title"] = [title]
    audio["artist"] = [artist]
    audio["album"] = [album]
    audio["lyrics"] = [lyrics]
    audio.clear_pictures()
    audio.add_picture(picture())
    audio.save()
elif fmt == "ogg":
    audio = OggVorbis(path)
    audio.clear()
    audio["title"] = [title]
    audio["artist"] = [artist]
    audio["album"] = [album]
    audio["lyrics"] = [lyrics]
    audio["metadata_block_picture"] = [base64.b64encode(picture().write()).decode("ascii")]
    audio.save()
else:
    raise SystemExit("unsupported format: " + fmt)
`

const mutagenReadScript = `
import base64
import json
import sys

from mutagen.flac import FLAC, Picture
from mutagen.id3 import APIC, ID3, TALB, TIT2, TPE1, TXXX, USLT
from mutagen.mp4 import MP4, MP4Cover
from mutagen.oggvorbis import OggVorbis
from mutagen.wave import WAVE

cfg = json.load(open(sys.argv[1], "r", encoding="utf-8"))
fmt = cfg["format"]
path = cfg["path"]
out = {"title": "", "artist": "", "album": "", "lyrics": "", "cover_mime": "", "cover_len": 0}

def first_text(tags, key):
    frames = tags.getall(key)
    if not frames:
        return ""
    text = getattr(frames[0], "text", "")
    if isinstance(text, list):
        return str(text[0]) if text else ""
    return str(text)

def read_id3(tags):
    out["title"] = first_text(tags, "TIT2")
    out["artist"] = first_text(tags, "TPE1")
    out["album"] = first_text(tags, "TALB")
    uslt = tags.getall("USLT")
    if uslt:
        out["lyrics"] = uslt[0].text
    else:
        for frame in tags.getall("TXXX"):
            if frame.desc in ("lyrics", "USLT"):
                out["lyrics"] = frame.text[0] if isinstance(frame.text, list) else frame.text
                break
    apic = tags.getall("APIC")
    if apic:
        out["cover_mime"] = apic[0].mime
        out["cover_len"] = len(apic[0].data)

if fmt in ("mp3", "aac"):
    read_id3(ID3(path))
elif fmt == "wav":
    audio = WAVE(path)
    read_id3(audio.tags)
elif fmt == "m4a":
    audio = MP4(path)
    out["title"] = audio.get("\xa9nam", [""])[0]
    out["artist"] = audio.get("\xa9ART", [""])[0]
    out["album"] = audio.get("\xa9alb", [""])[0]
    out["lyrics"] = audio.get("\xa9lyr", [""])[0]
    covr = audio.get("covr", [])
    if covr:
        out["cover_len"] = len(covr[0])
        out["cover_mime"] = "image/jpeg" if covr[0].imageformat == MP4Cover.FORMAT_JPEG else "image/png"
elif fmt == "flac":
    audio = FLAC(path)
    out["title"] = audio.get("title", [""])[0]
    out["artist"] = audio.get("artist", [""])[0]
    out["album"] = audio.get("album", [""])[0]
    out["lyrics"] = audio.get("lyrics", [""])[0]
    if audio.pictures:
        out["cover_mime"] = audio.pictures[0].mime
        out["cover_len"] = len(audio.pictures[0].data)
elif fmt == "ogg":
    audio = OggVorbis(path)
    out["title"] = audio.get("title", [""])[0]
    out["artist"] = audio.get("artist", [""])[0]
    out["album"] = audio.get("album", [""])[0]
    out["lyrics"] = audio.get("lyrics", [""])[0]
    pics = audio.get("metadata_block_picture", [])
    if pics:
        pic = Picture(base64.b64decode(pics[0]))
        out["cover_mime"] = pic.mime
        out["cover_len"] = len(pic.data)
else:
    raise SystemExit("unsupported format: " + fmt)

print(json.dumps(out, ensure_ascii=False))
`
