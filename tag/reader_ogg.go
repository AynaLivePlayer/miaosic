package tag

import "io"

func ReadOGGTags(r io.ReadSeeker) (Metadata, error) {
	return fallbackRead(r)
}
