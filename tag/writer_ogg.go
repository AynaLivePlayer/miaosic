package tag

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

var oggCRCTable = makeOggCRCTable()

type oggPage struct {
	headerType byte
	granule    uint64
	serial     uint32
	seq        uint32
	data       []byte
}

type oggPacket struct {
	data       []byte
	headerType byte
	granule    uint64
	startPage  int
	endPage    int
}

func WriteOGGTags(f *os.File, meta Metadata) error {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	pages, err := parseOggPages(data)
	if err != nil {
		return err
	}
	if len(pages) == 0 {
		return fmt.Errorf("miaosic: empty ogg file")
	}

	packets, err := parseOggPackets(pages)
	if err != nil {
		return err
	}
	if len(packets) < 2 {
		return fmt.Errorf("miaosic: ogg comment packet not found")
	}

	headerPackets, err := oggHeaderPacketCount(packets)
	if err != nil {
		return err
	}
	if len(packets) < headerPackets {
		return fmt.Errorf("miaosic: ogg header packet not found")
	}

	comment, err := buildOggCommentPacket(packets[1].data, meta)
	if err != nil {
		return err
	}
	packets[1].data = comment

	var out bytes.Buffer
	seq := uint32(0)
	for i, packet := range packets[:headerPackets] {
		headerType := packet.headerType &^ 0x01
		if i == 0 {
			headerType = 0x02
		}
		newPages, err := encodeOggPacketPages(packet.data, pages[0].serial, seq, headerType, 0)
		if err != nil {
			return err
		}
		for _, page := range newPages {
			out.Write(page)
			seq++
		}
	}

	headerEndPage := packets[headerPackets-1].endPage
	firstPreservedPage := headerEndPage + 1
	for i := headerPackets; i < len(packets) && packets[i].startPage == headerEndPage; i++ {
		packet := packets[i]
		newPages, err := encodeOggPacketPages(packet.data, pages[0].serial, seq, packet.headerType&^0x03, packet.granule)
		if err != nil {
			return err
		}
		for _, page := range newPages {
			out.Write(page)
			seq++
		}
		if packet.endPage >= firstPreservedPage {
			firstPreservedPage = packet.endPage + 1
		}
	}

	for _, page := range pages[firstPreservedPage:] {
		p := append([]byte(nil), page.data...)
		p[5] &^= 0x02
		binary.LittleEndian.PutUint32(p[18:22], seq)
		putOggCRC(p)
		out.Write(p)
		seq++
	}

	if err = f.Truncate(0); err != nil {
		return err
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	_, err = f.Write(out.Bytes())
	return err
}

func parseOggPackets(pages []oggPage) ([]oggPacket, error) {
	var packets []oggPacket
	var current []byte
	var headerType byte
	startPage := -1
	for pageIdx, page := range pages {
		segments := int(page.data[26])
		payload := page.data[27+segments:]
		pos := 0
		for _, lace := range page.data[27 : 27+segments] {
			size := int(lace)
			if pos+size > len(payload) {
				return nil, fmt.Errorf("miaosic: invalid ogg segment table")
			}
			if current == nil {
				headerType = page.headerType
				startPage = pageIdx
			}
			current = append(current, payload[pos:pos+size]...)
			pos += size
			if lace < 255 {
				packets = append(packets, oggPacket{
					data:       current,
					headerType: headerType,
					granule:    page.granule,
					startPage:  startPage,
					endPage:    pageIdx,
				})
				current = nil
				headerType = 0
				startPage = -1
			}
		}
	}
	if len(current) > 0 {
		return nil, fmt.Errorf("miaosic: unfinished ogg packet")
	}
	return packets, nil
}

func oggHeaderPacketCount(packets []oggPacket) (int, error) {
	switch {
	case len(packets) >= 2 && bytes.HasPrefix(packets[0].data, []byte("OpusHead")) && bytes.HasPrefix(packets[1].data, []byte("OpusTags")):
		return 2, nil
	case len(packets) >= 3 && bytes.HasPrefix(packets[0].data, []byte("\x01vorbis")) && bytes.HasPrefix(packets[1].data, []byte("\x03vorbis")):
		return 3, nil
	default:
		return 0, fmt.Errorf("miaosic: unsupported ogg header packets")
	}
}

func parseOggPages(data []byte) ([]oggPage, error) {
	var pages []oggPage
	for off := 0; off < len(data); {
		if off+27 > len(data) || string(data[off:off+4]) != "OggS" {
			return nil, fmt.Errorf("miaosic: invalid ogg page")
		}
		segments := int(data[off+26])
		if off+27+segments > len(data) {
			return nil, fmt.Errorf("miaosic: invalid ogg page segments")
		}
		payloadLen := 0
		for _, lace := range data[off+27 : off+27+segments] {
			payloadLen += int(lace)
		}
		end := off + 27 + segments + payloadLen
		if end > len(data) {
			return nil, fmt.Errorf("miaosic: truncated ogg page")
		}
		pageData := append([]byte(nil), data[off:end]...)
		pages = append(pages, oggPage{
			headerType: pageData[5],
			granule:    binary.LittleEndian.Uint64(pageData[6:14]),
			serial:     binary.LittleEndian.Uint32(pageData[14:18]),
			seq:        binary.LittleEndian.Uint32(pageData[18:22]),
			data:       pageData,
		})
		off = end
	}
	return pages, nil
}

func buildOggCommentPacket(original []byte, meta Metadata) ([]byte, error) {
	switch {
	case bytes.HasPrefix(original, []byte("\x03vorbis")):
		return buildVorbisCommentPacket([]byte("\x03vorbis"), true, meta), nil
	case bytes.HasPrefix(original, []byte("OpusTags")):
		return buildVorbisCommentPacket([]byte("OpusTags"), false, meta), nil
	default:
		return nil, fmt.Errorf("miaosic: unsupported ogg comment packet")
	}
}

func buildVorbisCommentPacket(prefix []byte, framingBit bool, meta Metadata) []byte {
	comments := []string{}
	add := func(k, v string) {
		if v != "" {
			comments = append(comments, strings.ToUpper(k)+"="+v)
		}
	}
	add("title", meta.Title)
	add("artist", meta.Artist)
	add("album", meta.Album)
	if len(meta.Lyrics) > 0 {
		add("lyrics", meta.Lyrics[0].Lyrics)
	}
	if len(meta.Pictures) > 0 {
		add("metadata_block_picture", vorbisPictureBlock(meta.Pictures[0]))
	}

	var buf bytes.Buffer
	buf.Write(prefix)
	writeLE32(&buf, uint32(len("miaosic")))
	buf.WriteString("miaosic")
	writeLE32(&buf, uint32(len(comments)))
	for _, comment := range comments {
		writeLE32(&buf, uint32(len(comment)))
		buf.WriteString(comment)
	}
	if framingBit {
		buf.WriteByte(1)
	}
	return buf.Bytes()
}

func encodeOggPacketPages(packet []byte, serial, seq uint32, headerType byte, granule uint64) ([][]byte, error) {
	var pages [][]byte
	pos := 0
	first := true
	for pos < len(packet) || first {
		remaining := len(packet) - pos
		payloadLen := remaining
		complete := true
		if payloadLen > 255*255 {
			payloadLen = 255 * 255
			complete = false
		}
		laces := make([]byte, 0, 255)
		for n := payloadLen; n >= 255; n -= 255 {
			laces = append(laces, 255)
		}
		if complete {
			laces = append(laces, byte(payloadLen%255))
		}
		if len(laces) > 255 {
			return nil, fmt.Errorf("miaosic: ogg page has too many lacing values")
		}

		pageHeaderType := headerType
		if !first {
			pageHeaderType = 0x01
		}
		pageGranule := uint64(^uint64(0))
		if complete {
			pageGranule = granule
		}

		var buf bytes.Buffer
		buf.WriteString("OggS")
		buf.WriteByte(0)
		buf.WriteByte(pageHeaderType)
		var tmp [8]byte
		binary.LittleEndian.PutUint64(tmp[:], pageGranule)
		buf.Write(tmp[:])
		binary.LittleEndian.PutUint32(tmp[:4], serial)
		buf.Write(tmp[:4])
		binary.LittleEndian.PutUint32(tmp[:4], seq+uint32(len(pages)))
		buf.Write(tmp[:4])
		buf.Write([]byte{0, 0, 0, 0})
		buf.WriteByte(byte(len(laces)))
		buf.Write(laces)
		buf.Write(packet[pos : pos+payloadLen])
		page := buf.Bytes()
		putOggCRC(page)
		pages = append(pages, page)
		pos += payloadLen
		first = false
	}
	return pages, nil
}

func putOggCRC(page []byte) {
	page[22], page[23], page[24], page[25] = 0, 0, 0, 0
	sum := oggChecksum(page)
	binary.LittleEndian.PutUint32(page[22:26], sum)
}

func makeOggCRCTable() [256]uint32 {
	var table [256]uint32
	for i := range table {
		r := uint32(i) << 24
		for range 8 {
			if r&0x80000000 != 0 {
				r = (r << 1) ^ 0x04c11db7
			} else {
				r <<= 1
			}
		}
		table[i] = r
	}
	return table
}

func oggChecksum(data []byte) uint32 {
	var crc uint32
	for _, b := range data {
		crc = (crc << 8) ^ oggCRCTable[byte(crc>>24)^b]
	}
	return crc
}

func writeLE32(buf *bytes.Buffer, v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	buf.Write(b[:])
}
