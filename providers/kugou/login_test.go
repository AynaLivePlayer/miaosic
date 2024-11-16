package kugou

import (
	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSignatureWebParams(t *testing.T) {
	params := map[string]string{
		"appid":      "1014",
		"clientver":  "8131",
		"clienttime": "1731737025938",
		"dfid":       "2C6khO0EyplK21gnln0Kok3i",
		"mid":        "ea502fad361e90cbe930fe6ea5d4dad2",
		"plat":       "4",
		"qrcode_txt": "https://h5.kugou.com/apps/loginQRCode/html/index.html?appid=1014&",
		"srcappid":   "2919",
		"type":       "1",
		"uuid":       "ea502fad361e90cbe930fe6ea5d4dad2",
	}

	expectedSignature := "FAC64F6FEDEBCAA4B41A2B1F22D1B5C9"
	signature := signatureWebParams(params)
	require.Equal(t, expectedSignature, signature)
}

func TestKugou_QrLogin(t *testing.T) {
	login, err := api.QrLogin()
	require.NoError(t, err, "QrLogin Error")
	require.NotEmpty(t, login, "QrLogin Result Empty")
	t.Log(login)
}

func TestKugou_QrLoginVerify(t *testing.T) {
	result, err := api.QrLoginVerify(&miaosic.QrLoginSession{
		Url: "https://h5.kugou.com/apps/loginQRCode/html/index.html?qrcode=5526fbc7576759da17bdeea9f02269323116",
		Key: "5526fbc7576759da17bdeea9f02269323116",
	})
	require.NoError(t, err, "QrLoginVerify Error")
	require.NotEmpty(t, result, "QrLoginVerify Result Empty")
	t.Log(result)
	if result.Success {
		t.Log("session", api.SaveSession())
	}
}
