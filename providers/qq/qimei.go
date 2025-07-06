package qq

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"strings"
	"time"
)

const (
	QiMeiPublicKey = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDEIxgwoutfwoJxcGQeedgP7FG9qaIuS0qzfR8gWkrkTZKM2iWHn2ajQpBRZjMSoSf6+KJGvar2ORhBfpDXyVtZCKpqLQ+FLkpncClKVIrBwv6PHyUvuCb0rIarmgDnzkfQAqVufEtR64iazGDKatvJ9y6B9NMbHddGSAUmRTCrHQIDAQAB
-----END PUBLIC KEY-----`
	QiMeiSecret = "ZdJqM15EeO2zWc08"
	QiMeiAppKey = "0AND0HD6FE4HY80F"
)

type QimeiResult struct {
	Q16 string `json:"q16"`
	Q36 string `json:"q36"`
}

type qimeiReserved struct {
	Harmony    string `json:"harmony"`
	Clone      string `json:"clone"`
	Containe   string `json:"containe"`
	Oz         string `json:"oz"`
	Oo         string `json:"oo"`
	Kelong     string `json:"kelong"`
	Uptimes    string `json:"uptimes"`
	MultiUser  string `json:"multiUser"`
	Bod        string `json:"bod"`
	Dv         string `json:"dv"`
	FirstLevel string `json:"firstLevel"`
	Manufact   string `json:"manufact"`
	Name       string `json:"name"`
	Host       string `json:"host"`
	Kernel     string `json:"kernel"`
}

type qimeiPayload struct {
	AndroidID        string `json:"androidId"`
	PlatformID       int    `json:"platformId"`
	AppKey           string `json:"appKey"`
	AppVersion       string `json:"appVersion"`
	BeaconIDSrc      string `json:"beaconIdSrc"`
	Brand            string `json:"brand"`
	ChannelID        string `json:"channelId"`
	Cid              string `json:"cid"`
	Imei             string `json:"imei"`
	Imsi             string `json:"imsi"`
	Mac              string `json:"mac"`
	Model            string `json:"model"`
	NetworkType      string `json:"networkType"`
	Oaid             string `json:"oaid"`
	OSVersion        string `json:"osVersion"`
	Qimei            string `json:"qimei"`
	Qimei36          string `json:"qimei36"`
	SDKVersion       string `json:"sdkVersion"`
	TargetSDKVersion string `json:"targetSdkVersion"`
	Audit            string `json:"audit"`
	UserID           string `json:"userId"`
	PackageID        string `json:"packageId"`
	DeviceType       string `json:"deviceType"`
	SDKName          string `json:"sdkName"`
	Reserved         string `json:"reserved"`
}

type qimeiParams struct {
	Key    string `json:"key"`
	Params string `json:"params"`
	Time   string `json:"time"`
	Nonce  string `json:"nonce"`
	Sign   string `json:"sign"`
	Extra  string `json:"extra"`
}

type qimeiRequest struct {
	App         int         `json:"app"`
	Os          int         `json:"os"`
	QimeiParams qimeiParams `json:"qimeiParams"`
}

type qimeiResponse struct {
	Data string `json:"data"`
}

type qimeiData struct {
	Data struct {
		Q16 string `json:"q16"`
		Q36 string `json:"q36"`
	} `json:"data"`
}

func qimeiRandomBeaconID() string {
	timeMonth := time.Now().Format("2006-01-") + "01"
	rand1 := rng.Intn(900000) + 100000
	rand2 := rng.Intn(900000000) + 100000000

	var parts []string
	for i := 1; i <= 40; i++ {
		switch i {
		case 1, 2, 13, 14, 17, 18, 21, 22, 25, 26, 29, 30, 33, 34, 37, 38:
			parts = append(parts, fmt.Sprintf("k%d:%s%d.%d", i, timeMonth, rand1, rand2))
		case 3:
			parts = append(parts, "k3:0000000000000000")
		case 4:
			// 生成16位随机十六进制
			buf := make([]byte, 8)
			rng.Read(buf)
			parts = append(parts, fmt.Sprintf("k4:%x", buf))
		default:
			parts = append(parts, fmt.Sprintf("k%d:%d", i, rng.Intn(10000)))
		}
	}
	return strings.Join(parts, ";")
}

func qimeiRandomPayloadByDevice(device *Device, version string) *qimeiPayload {
	fixedRand := rng.Intn(14400)
	uptime := time.Now().Add(-time.Duration(fixedRand) * time.Second).Format("2006-01-02 15:04:05")

	reserved := &qimeiReserved{
		Harmony:    "0",
		Clone:      "0",
		Containe:   "",
		Oz:         "UhYmelwouA+V2nPWbOvLTgN2/m8jwGB+yUB5v9tysQg=",
		Oo:         "Xecjt+9S1+f8Pz2VLSxgpw==",
		Kelong:     "0",
		Uptimes:    uptime,
		MultiUser:  "0",
		Bod:        device.Brand,
		Dv:         device.Device,
		FirstLevel: "",
		Manufact:   device.Brand,
		Name:       device.Model,
		Host:       "se.infra",
		Kernel:     device.ProcVersion,
	}

	reservedJSON, _ := json.Marshal(reserved)

	return &qimeiPayload{
		AndroidID:        device.AndroidID,
		PlatformID:       1,
		AppKey:           QiMeiAppKey,
		AppVersion:       version,
		BeaconIDSrc:      qimeiRandomBeaconID(),
		Brand:            device.Brand,
		ChannelID:        "10003505",
		Cid:              "",
		Imei:             device.IMEI,
		Imsi:             "",
		Mac:              "",
		Model:            device.Model,
		NetworkType:      "unknown",
		Oaid:             "",
		OSVersion:        fmt.Sprintf("Android %s,level %d", device.Version.Release, device.Version.Sdk),
		Qimei:            "",
		Qimei36:          "",
		SDKVersion:       version,
		TargetSDKVersion: "33",
		Audit:            "",
		UserID:           "{}",
		PackageID:        "com.tencent.qqmusic",
		DeviceType:       "Phone",
		SDKName:          "",
		Reserved:         string(reservedJSON),
	}
}

func getQimei(device *Device, version string) (*QimeiResult, error) {
	result, err := fetchQimei(device, version)
	if err == nil {
		device.Qimei = result.Q36
		return result, nil
	}
	if device.Qimei != "" {
		return &QimeiResult{Q16: "", Q36: device.Qimei}, nil
	}
	return &QimeiResult{Q16: "", Q36: "6c9d3cd110abca9b16311cee10001e717614"}, nil
}

func qimeiRandomString(length int) string {
	const charset = "abcdef0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rng.Intn(len(charset))]
	}
	return string(result)
}

// 从腾讯API获取QIMEI
func fetchQimei(device *Device, version string) (*QimeiResult, error) {
	payload := qimeiRandomPayloadByDevice(device, version)
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 生成加密密钥和随机数
	cryptKey := qimeiRandomString(16)
	nonce := qimeiRandomString(16)
	ts := time.Now().UnixMilli()

	// RSA加密密钥
	rsaEncryptedKey, err := rsaEncrypt([]byte(cryptKey), QiMeiPublicKey)
	if err != nil {
		return nil, err
	}
	keyBase64 := base64.StdEncoding.EncodeToString(rsaEncryptedKey)

	// AES加密payload
	aesEncrypted, err := aesEncrypt([]byte(cryptKey), payloadJSON)
	if err != nil {
		return nil, err
	}
	paramsBase64 := base64.StdEncoding.EncodeToString(aesEncrypted)

	// 生成签名
	extra := fmt.Sprintf(`{"appKey":"%s"}`, QiMeiAppKey)
	sign := calcMd5(keyBase64, paramsBase64, fmt.Sprintf("%d", ts), nonce, QiMeiSecret, extra)

	// 构建请求
	qimeiParams := qimeiParams{
		Key:    keyBase64,
		Params: paramsBase64,
		Time:   fmt.Sprintf("%d", ts),
		Nonce:  nonce,
		Sign:   sign,
		Extra:  extra,
	}

	requestBody := qimeiRequest{
		App:         0,
		Os:          1,
		QimeiParams: qimeiParams,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := miaosic.Requester.Post(
		"https://api.tencentmusic.com/tme/trpc/proxy",
		map[string]string{
			"Host":         "api.tencentmusic.com",
			"method":       "GetQimei",
			"service":      "trpc.tme_datasvr.qimeiproxy.QimeiProxy",
			"appid":        "qimei_qq_android",
			"sign":         calcMd5("qimei_qq_androidpzAuCmaFAaFaHrdakPjLIEqKrGnSOOvH", fmt.Sprintf("%d", ts/1000)),
			"user-agent":   "QQMusic",
			"timestamp":    fmt.Sprintf("%d", ts/1000),
			"Content-Type": "application/json",
		}, requestJSON,
	)
	if err != nil {
		return nil, err
	}

	var qimeiResp qimeiResponse
	if err := json.Unmarshal(resp.Body(), &qimeiResp); err != nil {
		return nil, err
	}

	var qimeiData qimeiData
	if err := json.Unmarshal([]byte(qimeiResp.Data), &qimeiData); err != nil {
		return nil, err
	}

	return &QimeiResult{
		Q16: qimeiData.Data.Q16,
		Q36: qimeiData.Data.Q36,
	}, nil
}
