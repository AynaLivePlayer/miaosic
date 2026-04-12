package tag

import (
	"bytes"
	"errors"
	"github.com/gabriel-vasile/mimetype"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func fixMeta(meta *Metadata) {
	// fix picture meme
	for idx, _ := range meta.Pictures {
		if meta.Pictures[idx].Mimetype == "" {
			meta.Pictures[idx].Mimetype = mimetype.Detect(meta.Pictures[idx].Data).String()
		}
	}
}

// Write metadata to file, input file will be closed after this method
func Write(f *os.File, meta Metadata) error {
	defer f.Close()
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	b := make([]byte, 512)
	n, err := f.Read(b)
	if err != nil && err != io.EOF {
		return err
	}
	b = b[:n]
	mimeType := mimetype.Detect(b).String()
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	fixMeta(&meta)
	switch {
	case len(b) >= 4 && string(b[0:4]) == "fLaC":
		return WriteFlacTags(f, meta)
	case len(b) >= 12 && string(b[0:4]) == "RIFF" && string(b[8:12]) == "WAVE":
		return WriteWAVTags(f, meta)
	case len(b) >= 4 && bytes.Equal(b[0:4], []byte("OggS")):
		return WriteOGGTags(f, meta)
	case len(b) >= 8 && string(b[4:8]) == "ftyp":
		return WriteMP4Tags(f, meta)
	case mimeType == "audio/mpeg" || isLikelyAAC(f.Name(), b) || (len(b) >= 3 && string(b[0:3]) == "ID3"):
		return WriteID3v2Tags(f, meta)
	}
	return errors.New("miaosic: mime-type not supported")
}

func WriteTo(path string, meta Metadata) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	return Write(file, meta)
}

func isLikelyAAC(name string, b []byte) bool {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".aac" || ext == ".acc" {
		return true
	}
	return len(b) >= 2 && b[0] == 0xff && b[1]&0xf6 == 0xf0
}
