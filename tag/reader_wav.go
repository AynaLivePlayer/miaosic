package tag

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"

	"github.com/bogem/id3v2/v2"
)

func ReadWAVTags(r io.ReadSeeker) (Metadata, error) {
	mime, err := detectMime(r)
	if err != nil {
		return Metadata{}, err
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return Metadata{}, err
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return Metadata{}, err
	}
	meta := Metadata{
		Format:   FormatRIFF,
		Mimetype: mime,
		Lyrics:   []Lyrics{},
		Pictures: []Picture{},
	}
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return meta, errors.New("miaosic: not a wav file")
	}

	for off := 12; off+8 <= len(data); {
		id := string(data[off : off+4])
		size := int(binary.LittleEndian.Uint32(data[off+4 : off+8]))
		payloadStart := off + 8
		payloadEnd := payloadStart + size
		if size < 0 || payloadEnd > len(data) {
			break
		}
		payload := data[payloadStart:payloadEnd]
		switch id {
		case "LIST":
			readWAVInfo(payload, &meta)
		case "id3 ", "ID3 ":
			readWAVID3(payload, &meta)
		}
		off = payloadEnd
		if size%2 == 1 {
			off++
		}
	}
	return meta, nil
}

func readWAVInfo(payload []byte, meta *Metadata) {
	if len(payload) < 4 || string(payload[:4]) != "INFO" {
		return
	}
	for off := 4; off+8 <= len(payload); {
		id := string(payload[off : off+4])
		size := int(binary.LittleEndian.Uint32(payload[off+4 : off+8]))
		start := off + 8
		end := start + size
		if size < 0 || end > len(payload) {
			return
		}
		value := strings.TrimRight(string(payload[start:end]), "\x00")
		switch id {
		case "INAM":
			meta.Title = value
		case "IART":
			meta.Artist = value
		case "IPRD":
			meta.Album = value
		case "ILYR":
			if value != "" {
				meta.Lyrics = append(meta.Lyrics, Lyrics{Lang: "unk", Lyrics: value})
			}
		}
		off = end
		if size%2 == 1 {
			off++
		}
	}
}

func readWAVID3(payload []byte, meta *Metadata) {
	tag, err := id3v2.ParseReader(bytes.NewReader(payload), id3v2.Options{Parse: true})
	if err != nil {
		return
	}
	if tag.Title() != "" {
		meta.Title = tag.Title()
	}
	if tag.Artist() != "" {
		meta.Artist = tag.Artist()
	}
	if tag.Album() != "" {
		meta.Album = tag.Album()
	}
	for _, frame := range tag.GetFrames("USLT") {
		lyricFrame, ok := frame.(id3v2.UnsynchronisedLyricsFrame)
		if !ok {
			continue
		}
		meta.Lyrics = append(meta.Lyrics, Lyrics{
			Lang:   lyricFrame.Language,
			Lyrics: lyricFrame.Lyrics,
		})
	}
	for _, frame := range tag.GetFrames("APIC") {
		pic, ok := frame.(id3v2.PictureFrame)
		if !ok {
			continue
		}
		meta.Pictures = append(meta.Pictures, Picture{
			Mimetype:    pic.MimeType,
			Type:        pic.PictureType,
			Description: pic.Description,
			Data:        pic.Picture,
		})
	}
}
