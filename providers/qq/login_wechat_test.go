package qq

import (
	"github.com/AynaLivePlayer/miaosic"
	"github.com/k0kubun/pp/v3"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQQ_getQrcodeWx(t *testing.T) {
	result, err := testApi.getWxQR()
	require.NoError(t, err)
	require.NotEmpty(t, result)
	pp.Println(result)
}

func TestQQ_checkWxQR(t *testing.T) {
	_, err := testApi.checkWxQR(&miaosic.QrLoginSession{
		Url: "https://open.weixin.qq.com/connect/confirm?uuid=071Yrpg10iD00w3H",
		Key: "071Yrpg10iD00w3H",
	})
	require.NoError(t, err)
	//pp.Println(result)
}
