package qq

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQQ_getQrcodeQQ(t *testing.T) {
	_, err := testApi.getQQQR()
	require.NoError(t, err)
	//pp.Println(result)
}
