package tag

import (
	"github.com/k0kubun/pp/v3"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestReader(t *testing.T) {
	f, err := os.Open("/home/aynakeya/workspace/AynaLivePlayer/pkg/miaosic/cmd/miaosic/Mili - world.execute (me) ;.mp3")
	require.NoError(t, err)
	meta, err := Read(f)
	require.NoError(t, err)
	pp.Println(meta)
	f.Close()
}

func TestReader_Flac(t *testing.T) {
	f, err := os.Open("/home/aynakeya/workspace/AynaLivePlayer/pkg/miaosic/cmd/miaosic/欢子 - 心痛2009.flac")
	require.NoError(t, err)
	meta, err := Read(f)
	require.NoError(t, err)
	pp.Println(meta)
	f.Close()
}
