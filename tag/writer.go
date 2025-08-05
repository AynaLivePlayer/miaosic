package tag

import (
	"errors"
	"github.com/gabriel-vasile/mimetype"
	"io"
	"os"
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
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	b := make([]byte, 512)
	_, err = io.ReadFull(f, b)
	if err != nil {
		return err
	}
	mimeType := mimetype.Detect(b).String()
	_, err = f.Seek(0, io.SeekStart)
	fixMeta(&meta)
	switch mimeType {
	case "audio/flac", "audio/x-flac":
		return WriteFlacTags(f, meta)
	//case string(b[0:4]) == "OggS":
	//	return ReadOGGTags(r)
	//case string(b[4:8]) == "ftyp":
	//	return ReadAtoms(r)
	case "audio/mpeg":
		return WriteID3v2Tags(f, meta)
		//case string(b[0:4]) == "DSD ":
		//	return ReadDSFTags(r)
	}
	return errors.New("miaosic: mime-type not supported")
}

func WriteTo(path string, meta Metadata) error {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return err
	}
	return Write(file, meta)
}
