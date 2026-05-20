package tag

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type testOggPacket struct {
	data    []byte
	granule uint64
}

func TestWriteOGGTagsPreservesOpusAudioGranule(t *testing.T) {
	// Opus only has two header packets. This protects against treating the first
	// audio packet as a third header packet and rewriting its granule position to 0.
	path := filepath.Join(t.TempDir(), "sample.opus")
	original := bytes.Join([][]byte{
		testOggPage(t, 1, 0, 0x02, 0, []byte("OpusHead")),
		testOggPage(t, 1, 1, 0x00, 0, []byte("OpusTags")),
		testOggPage(t, 1, 2, 0x00, 960, []byte("opus-audio-1")),
		testOggPage(t, 1, 3, 0x00, 1920, []byte("opus-audio-2")),
	}, nil)
	require.NoError(t, os.WriteFile(path, original, 0o600))

	require.NoError(t, WriteTo(path, Metadata{Title: "new title"}))

	packets := readTestOggPackets(t, path)
	require.Len(t, packets, 4)
	require.Equal(t, []byte("opus-audio-1"), packets[2].data)
	require.Equal(t, uint64(960), packets[2].granule)
	require.Equal(t, []byte("opus-audio-2"), packets[3].data)
	require.Equal(t, uint64(1920), packets[3].granule)
}

func TestWriteOGGTagsKeepsAudioPacketSharingVorbisHeaderPage(t *testing.T) {
	// Vorbis has three header packets, but an audio packet can share the same page
	// as the third header packet. This protects against dropping that same-page audio packet.
	path := filepath.Join(t.TempDir(), "sample.ogg")
	original := bytes.Join([][]byte{
		testOggPage(t, 2, 0, 0x02, 0, []byte("\x01vorbis")),
		testOggPage(t, 2, 1, 0x00, 0, []byte("\x03vorbis")),
		testOggPage(t, 2, 2, 0x00, 960, []byte("\x05vorbis"), []byte("audio-same-page")),
		testOggPage(t, 2, 3, 0x00, 1920, []byte("audio-next-page")),
	}, nil)
	require.NoError(t, os.WriteFile(path, original, 0o600))

	require.NoError(t, WriteTo(path, Metadata{Title: "new title"}))

	packets := readTestOggPackets(t, path)
	require.Contains(t, testOggPacketStrings(packets), "audio-same-page")
	require.Contains(t, testOggPacketStrings(packets), "audio-next-page")
}

func TestWriteWAVTagsPreservesNonInfoListChunks(t *testing.T) {
	// WAV LIST chunks are not always metadata. This protects non-INFO LIST chunks
	// such as adtl from being deleted when replacing tags.
	path := filepath.Join(t.TempDir(), "sample.wav")
	adtl := append([]byte("adtl"), []byte("associated data")...)
	original := testRIFF("WAVE",
		testRIFFChunk("fmt ", make([]byte, 16)),
		testRIFFChunk("LIST", adtl),
		testRIFFChunk("data", []byte{1, 2, 3, 4}),
	)
	require.NoError(t, os.WriteFile(path, original, 0o600))

	require.NoError(t, WriteTo(path, Metadata{Title: "new title"}))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.True(t, hasRIFFListPayload(data, "adtl"), "non-INFO LIST chunk should be preserved")
}

func TestWriteFLACTagsClearsOldPicturesWhenNoPicturesProvided(t *testing.T) {
	// FLAC tag writing is overwrite-style. This protects against stale picture
	// blocks surviving when the new metadata intentionally has no pictures.
	requireFFmpeg(t)
	source, _, cover := testDataPaths(t)
	path := filepath.Join(t.TempDir(), "sample.flac")
	writeCleanSampleWithFFmpeg(t, tagFormatCase{name: "flac", ext: ".flac", args: []string{"-map", "0:a:0", "-map_metadata", "-1", "-t", "1", "-c:a", "flac"}}, source, path)

	meta := testMetadata("flac", cover)
	require.NoError(t, WriteTo(path, meta))
	got := readMetadata(t, path)
	require.NotEmpty(t, got.Pictures)

	meta.Pictures = nil
	require.NoError(t, WriteTo(path, meta))
	got = readMetadata(t, path)
	require.Empty(t, got.Pictures)
}

func testOggPage(t *testing.T, serial, seq uint32, headerType byte, granule uint64, packets ...[]byte) []byte {
	t.Helper()
	var laces []byte
	var payload bytes.Buffer
	for _, packet := range packets {
		require.Less(t, len(packet), 255)
		laces = append(laces, byte(len(packet)))
		payload.Write(packet)
	}

	var page bytes.Buffer
	page.WriteString("OggS")
	page.WriteByte(0)
	page.WriteByte(headerType)
	var tmp8 [8]byte
	binary.LittleEndian.PutUint64(tmp8[:], granule)
	page.Write(tmp8[:])
	var tmp4 [4]byte
	binary.LittleEndian.PutUint32(tmp4[:], serial)
	page.Write(tmp4[:])
	binary.LittleEndian.PutUint32(tmp4[:], seq)
	page.Write(tmp4[:])
	page.Write([]byte{0, 0, 0, 0})
	page.WriteByte(byte(len(laces)))
	page.Write(laces)
	page.Write(payload.Bytes())
	data := page.Bytes()
	putOggCRC(data)
	return data
}

func readTestOggPackets(t *testing.T, path string) []testOggPacket {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	pages, err := parseOggPages(data)
	require.NoError(t, err)
	var packets []testOggPacket
	var current []byte
	for _, page := range pages {
		segments := int(page.data[26])
		payload := page.data[27+segments:]
		pos := 0
		for _, lace := range page.data[27 : 27+segments] {
			size := int(lace)
			current = append(current, payload[pos:pos+size]...)
			pos += size
			if lace < 255 {
				packets = append(packets, testOggPacket{
					data:    append([]byte(nil), current...),
					granule: page.granule,
				})
				current = nil
			}
		}
	}
	return packets
}

func testOggPacketStrings(packets []testOggPacket) []string {
	out := make([]string, 0, len(packets))
	for _, packet := range packets {
		out = append(out, string(packet.data))
	}
	return out
}

func testRIFF(format string, chunks ...[]byte) []byte {
	var body bytes.Buffer
	body.WriteString(format)
	for _, chunk := range chunks {
		body.Write(chunk)
	}
	var out bytes.Buffer
	out.WriteString("RIFF")
	var size [4]byte
	binary.LittleEndian.PutUint32(size[:], uint32(body.Len()))
	out.Write(size[:])
	out.Write(body.Bytes())
	return out.Bytes()
}

func testRIFFChunk(id string, payload []byte) []byte {
	var out bytes.Buffer
	writeRIFFChunk(&out, id, payload)
	return out.Bytes()
}

func hasRIFFListPayload(data []byte, listType string) bool {
	for off := 12; off+12 <= len(data); {
		id := string(data[off : off+4])
		size := int(binary.LittleEndian.Uint32(data[off+4 : off+8]))
		start := off + 8
		end := start + size
		if end > len(data) {
			return false
		}
		if id == "LIST" && size >= 4 && string(data[start:start+4]) == listType {
			return true
		}
		off = end
		if size%2 == 1 {
			off++
		}
	}
	return false
}
