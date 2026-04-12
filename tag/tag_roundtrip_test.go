package tag

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTagRoundTripCommonFormats(t *testing.T) {
	requireFFmpeg(t)
	source, _, cover := testDataPaths(t)

	for _, tc := range tagTestFormats() {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "sample"+tc.ext)
			writeCleanSampleWithFFmpeg(t, tc, source, path)
			meta := testMetadata(tc.name, cover)

			require.NoError(t, WriteTo(path, meta))
			got := readMetadata(t, path)
			assertMetadata(t, meta, got, true, true)
		})
	}
}
