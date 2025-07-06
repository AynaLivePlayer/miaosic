package qq

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetQimei(t *testing.T) {
	device := NewDevice()
	qimei, err := getQimei(device, "13.2.5.8")
	require.NoError(t, err)
	require.NotEqual(t, "6c9d3cd110abca9b16311cee10001e717614", qimei.Q36)
}
