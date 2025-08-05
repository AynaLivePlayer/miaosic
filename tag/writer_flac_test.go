package tag

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestWriterFlac(t *testing.T) {
	f, err := os.Open("/home/aynakeya/workspace/AynaLivePlayer/pkg/miaosic/cmd/miaosic/data.flac")
	require.NoError(t, err)
	err = WriteFlacTags(f, Metadata{})
	require.NoError(t, err)
	require.NoError(t, f.Close())
}
