package qq

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQQ_getQrcodeQQ(t *testing.T) {
	_, err := testApi.getQQQR()
	require.NoError(t, err)
	//pp.Println(result)
}
