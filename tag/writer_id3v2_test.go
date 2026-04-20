package tag

import (
	"bytes"
	"testing"

	"github.com/bogem/id3v2/v2"
	"github.com/stretchr/testify/require"
)

func TestSetID3v2MetadataSupportsUnicode(t *testing.T) {
	meta := Metadata{
		Title:  "影子小姐",
		Artist: "封茗囧菌",
		Album:  "中文专辑",
		Lyrics: []Lyrics{{Lang: "zh", Lyrics: "这是一段中文歌词"}},
	}

	tag := id3v2.NewEmptyTag()
	setID3v2Metadata(tag, meta)

	var buf bytes.Buffer
	_, err := tag.WriteTo(&buf)
	require.NoError(t, err)
}
