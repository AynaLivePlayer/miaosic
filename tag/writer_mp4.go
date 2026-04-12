package tag

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type mp4Box struct {
	start      int
	size       int
	headerSize int
	typ        string
}

func WriteMP4Tags(f *os.File, meta Metadata) error {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	boxes, err := parseMP4Boxes(data, 0, len(data))
	if err != nil {
		return err
	}
	moovIdx := -1
	mdatStart := -1
	for i, box := range boxes {
		switch box.typ {
		case "moov":
			moovIdx = i
		case "mdat":
			mdatStart = box.start
		}
	}
	if moovIdx < 0 {
		return fmt.Errorf("miaosic: mp4 moov box not found")
	}

	moov := boxes[moovIdx]
	newMoov, err := rewriteMoov(data[moov.start:moov.start+moov.size], meta)
	if err != nil {
		return err
	}
	delta := len(newMoov) - moov.size
	if delta != 0 && mdatStart > moov.start {
		if err = adjustChunkOffsets(newMoov[8:], delta); err != nil {
			return err
		}
	}

	var out bytes.Buffer
	out.Write(data[:moov.start])
	out.Write(newMoov)
	out.Write(data[moov.start+moov.size:])

	if err = f.Truncate(0); err != nil {
		return err
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	_, err = f.Write(out.Bytes())
	return err
}

func rewriteMoov(moov []byte, meta Metadata) ([]byte, error) {
	children, err := parseMP4Boxes(moov, 8, len(moov))
	if err != nil {
		return nil, err
	}
	var payload bytes.Buffer
	for _, child := range children {
		if child.typ == "udta" {
			udta, keep, err := rewriteUdta(moov[child.start : child.start+child.size])
			if err != nil {
				return nil, err
			}
			if keep {
				payload.Write(udta)
			}
			continue
		}
		payload.Write(moov[child.start : child.start+child.size])
	}
	payload.Write(buildUdta(meta))
	return mp4BoxBytes("moov", payload.Bytes()), nil
}

func rewriteUdta(udta []byte) ([]byte, bool, error) {
	children, err := parseMP4Boxes(udta, 8, len(udta))
	if err != nil {
		return nil, false, err
	}
	var payload bytes.Buffer
	for _, child := range children {
		if child.typ != "meta" {
			payload.Write(udta[child.start : child.start+child.size])
		}
	}
	if payload.Len() == 0 {
		return nil, false, nil
	}
	return mp4BoxBytes("udta", payload.Bytes()), true, nil
}

func buildUdta(meta Metadata) []byte {
	var ilst bytes.Buffer
	writeMP4TextItem(&ilst, []byte{0xa9, 'n', 'a', 'm'}, meta.Title)
	writeMP4TextItem(&ilst, []byte{0xa9, 'A', 'R', 'T'}, meta.Artist)
	writeMP4TextItem(&ilst, []byte{0xa9, 'a', 'l', 'b'}, meta.Album)
	if len(meta.Lyrics) > 0 {
		writeMP4TextItem(&ilst, []byte{0xa9, 'l', 'y', 'r'}, meta.Lyrics[0].Lyrics)
	}
	if len(meta.Pictures) > 0 {
		writeMP4CoverItem(&ilst, meta.Pictures[0])
	}

	var metaPayload bytes.Buffer
	metaPayload.Write([]byte{0, 0, 0, 0})
	metaPayload.Write(buildMP4Handler())
	metaPayload.Write(mp4BoxBytes("ilst", ilst.Bytes()))
	return mp4BoxBytes("udta", mp4BoxBytes("meta", metaPayload.Bytes()))
}

func buildMP4Handler() []byte {
	var payload bytes.Buffer
	payload.Write([]byte{0, 0, 0, 0})
	payload.Write([]byte{0, 0, 0, 0})
	payload.WriteString("mdir")
	payload.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	payload.WriteByte(0)
	return mp4BoxBytes("hdlr", payload.Bytes())
}

func writeMP4TextItem(buf *bytes.Buffer, typ []byte, value string) {
	if value == "" {
		return
	}
	buf.Write(mp4TypedBoxBytes(typ, mp4DataBox(1, []byte(value))))
}

func writeMP4CoverItem(buf *bytes.Buffer, pic Picture) {
	dataType := uint32(13)
	if pic.Mimetype == "image/png" {
		dataType = 14
	}
	buf.Write(mp4TypedBoxBytes([]byte("covr"), mp4DataBox(dataType, pic.Data)))
}

func mp4DataBox(dataType uint32, data []byte) []byte {
	var payload bytes.Buffer
	writeBE32(&payload, dataType)
	writeBE32(&payload, 0)
	payload.Write(data)
	return mp4BoxBytes("data", payload.Bytes())
}

func mp4BoxBytes(typ string, payload []byte) []byte {
	return mp4TypedBoxBytes([]byte(typ), payload)
}

func mp4TypedBoxBytes(typ []byte, payload []byte) []byte {
	var buf bytes.Buffer
	writeBE32(&buf, uint32(len(payload)+8))
	buf.Write(typ[:4])
	buf.Write(payload)
	return buf.Bytes()
}

func parseMP4Boxes(data []byte, start, end int) ([]mp4Box, error) {
	var boxes []mp4Box
	for off := start; off < end; {
		if off+8 > end {
			return nil, fmt.Errorf("miaosic: invalid mp4 box")
		}
		size := int(binary.BigEndian.Uint32(data[off : off+4]))
		headerSize := 8
		if size == 1 {
			if off+16 > end {
				return nil, fmt.Errorf("miaosic: invalid mp4 largesize box")
			}
			size64 := binary.BigEndian.Uint64(data[off+8 : off+16])
			if size64 > uint64(^uint(0)>>1) {
				return nil, fmt.Errorf("miaosic: mp4 box too large")
			}
			size = int(size64)
			headerSize = 16
		} else if size == 0 {
			size = end - off
		}
		if size < headerSize || off+size > end {
			return nil, fmt.Errorf("miaosic: invalid mp4 box size")
		}
		boxes = append(boxes, mp4Box{
			start:      off,
			size:       size,
			headerSize: headerSize,
			typ:        string(data[off+4 : off+8]),
		})
		off += size
	}
	return boxes, nil
}

func adjustChunkOffsets(data []byte, delta int) error {
	boxes, err := parseMP4Boxes(data, 0, len(data))
	if err != nil {
		return err
	}
	for _, box := range boxes {
		payload := data[box.start+box.headerSize : box.start+box.size]
		switch box.typ {
		case "stco":
			if len(payload) < 8 {
				return fmt.Errorf("miaosic: invalid stco box")
			}
			count := int(binary.BigEndian.Uint32(payload[4:8]))
			if len(payload) < 8+count*4 {
				return fmt.Errorf("miaosic: invalid stco entries")
			}
			for i := 0; i < count; i++ {
				pos := 8 + i*4
				v := int64(binary.BigEndian.Uint32(payload[pos : pos+4]))
				v += int64(delta)
				if v < 0 || v > int64(^uint32(0)) {
					return fmt.Errorf("miaosic: stco offset out of range")
				}
				binary.BigEndian.PutUint32(payload[pos:pos+4], uint32(v))
			}
		case "co64":
			if len(payload) < 8 {
				return fmt.Errorf("miaosic: invalid co64 box")
			}
			count := int(binary.BigEndian.Uint32(payload[4:8]))
			if len(payload) < 8+count*8 {
				return fmt.Errorf("miaosic: invalid co64 entries")
			}
			for i := 0; i < count; i++ {
				pos := 8 + i*8
				v := int64(binary.BigEndian.Uint64(payload[pos : pos+8]))
				v += int64(delta)
				if v < 0 {
					return fmt.Errorf("miaosic: co64 offset out of range")
				}
				binary.BigEndian.PutUint64(payload[pos:pos+8], uint64(v))
			}
		case "moov", "trak", "mdia", "minf", "stbl", "edts", "dinf":
			if err = adjustChunkOffsets(payload, delta); err != nil {
				return err
			}
		}
	}
	return nil
}

func vorbisPictureBlock(pic Picture) string {
	var buf bytes.Buffer
	writeBE32(&buf, uint32(pic.Type))
	writeBE32(&buf, uint32(len(pic.Mimetype)))
	buf.WriteString(pic.Mimetype)
	writeBE32(&buf, uint32(len(pic.Description)))
	buf.WriteString(pic.Description)
	writeBE32(&buf, 0)
	writeBE32(&buf, 0)
	writeBE32(&buf, 0)
	writeBE32(&buf, 0)
	writeBE32(&buf, uint32(len(pic.Data)))
	buf.Write(pic.Data)
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func writeBE32(buf *bytes.Buffer, v uint32) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	buf.Write(b[:])
}
