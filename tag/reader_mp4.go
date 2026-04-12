package tag

import "io"

func ReadMP4Tags(r io.ReadSeeker) (Metadata, error) {
	return fallbackRead(r)
}
