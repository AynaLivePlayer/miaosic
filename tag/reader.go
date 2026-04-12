package tag

import (
	"bytes"
	"io"

	"github.com/gabriel-vasile/mimetype"
)

func Read(r io.ReadSeeker) (Metadata, error) {
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return Metadata{}, err
	}

	b := make([]byte, 16)
	_, err = r.Read(b)
	if err != nil && err != io.EOF {
		return Metadata{}, err
	}
	switch {
	case len(b) >= 4 && string(b[0:4]) == "fLaC":
		return ReadFLACTags(r)
	case len(b) >= 12 && string(b[0:4]) == "RIFF" && string(b[8:12]) == "WAVE":
		return ReadWAVTags(r)
	case len(b) >= 3 && string(b[0:3]) == "ID3":
		return ReadID3v2Tags(r)
	case len(b) >= 4 && bytes.Equal(b[0:4], []byte("OggS")):
		return ReadOGGTags(r)
	case len(b) >= 8 && string(b[4:8]) == "ftyp":
		return ReadMP4Tags(r)
	}
	return fallbackRead(r)
}

func detectMime(r io.ReadSeeker) (string, error) {
	pos, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return "", err
	}
	if _, err = r.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	b := make([]byte, 512)
	n, readErr := r.Read(b)
	if _, err = r.Seek(pos, io.SeekStart); err != nil {
		return "", err
	}
	if readErr != nil && readErr != io.EOF {
		return "", readErr
	}
	return mimetype.Detect(b[:n]).String(), nil
}
