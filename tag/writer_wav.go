package tag

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/bogem/id3v2/v2"
)

func WriteWAVTags(f *os.File, meta Metadata) error {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return fmt.Errorf("miaosic: not a wav file")
	}

	var out bytes.Buffer
	out.Write(data[:12])
	for off := 12; off+8 <= len(data); {
		id := string(data[off : off+4])
		size := int(binary.LittleEndian.Uint32(data[off+4 : off+8]))
		payloadEnd := off + 8 + size
		if size < 0 || payloadEnd > len(data) {
			break
		}
		next := payloadEnd
		if size%2 == 1 {
			next++
		}
		if id != "LIST" && id != "id3 " && id != "ID3 " {
			out.Write(data[off:next])
		}
		off = next
	}

	writeRIFFChunk(&out, "LIST", wavInfoPayload(meta))
	if id3Payload := wavID3Payload(meta); len(id3Payload) > 0 {
		writeRIFFChunk(&out, "id3 ", id3Payload)
	}
	result := out.Bytes()
	binary.LittleEndian.PutUint32(result[4:8], uint32(len(result)-8))

	if err = f.Truncate(0); err != nil {
		return err
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	_, err = f.Write(result)
	return err
}

func wavInfoPayload(meta Metadata) []byte {
	var payload bytes.Buffer
	payload.WriteString("INFO")
	writeRIFFChunk(&payload, "INAM", riffString(meta.Title))
	writeRIFFChunk(&payload, "IART", riffString(meta.Artist))
	writeRIFFChunk(&payload, "IPRD", riffString(meta.Album))
	if len(meta.Lyrics) > 0 {
		writeRIFFChunk(&payload, "ILYR", riffString(meta.Lyrics[0].Lyrics))
	}
	return payload.Bytes()
}

func wavID3Payload(meta Metadata) []byte {
	tag := id3v2.NewEmptyTag()
	setID3v2Metadata(tag, meta)
	var payload bytes.Buffer
	_, _ = tag.WriteTo(&payload)
	return payload.Bytes()
}

func writeRIFFChunk(buf *bytes.Buffer, id string, payload []byte) {
	buf.WriteString(id)
	var size [4]byte
	binary.LittleEndian.PutUint32(size[:], uint32(len(payload)))
	buf.Write(size[:])
	buf.Write(payload)
	if len(payload)%2 == 1 {
		buf.WriteByte(0)
	}
}

func riffString(value string) []byte {
	return append([]byte(value), 0)
}
