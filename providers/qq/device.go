package qq

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

// OSVersion 系统版本信息
type OSVersion struct {
	Incremental string `json:"incremental"`
	Release     string `json:"release"`
	Codename    string `json:"codename"`
	Sdk         int    `json:"sdk"`
}

func newOSVersion() *OSVersion {
	return &OSVersion{
		Incremental: "5891938",
		Release:     "10",
		Codename:    "REL",
		Sdk:         29,
	}
}

// Device 设备相关信息
type Device struct {
	Display      string     `json:"display"`
	Product      string     `json:"product"`
	Device       string     `json:"device"`
	Board        string     `json:"board"`
	Model        string     `json:"model"`
	Fingerprint  string     `json:"fingerprint"`
	BootID       string     `json:"boot_id"`
	ProcVersion  string     `json:"proc_version"`
	IMEI         string     `json:"imei"`
	Brand        string     `json:"brand"`
	Bootloader   string     `json:"bootloader"`
	BaseBand     string     `json:"base_band"`
	Version      *OSVersion `json:"version"`
	SimInfo      string     `json:"sim_info"`
	OsType       string     `json:"os_type"`
	MacAddress   string     `json:"mac_address"`
	WifiBSSID    string     `json:"wifi_bssid"`
	WifiSSID     string     `json:"wifi_ssid"`
	IMSIMD5      []byte     `json:"imsi_md5"`
	AndroidID    string     `json:"android_id"`
	APN          string     `json:"apn"`
	VendorName   string     `json:"vendor_name"`
	VendorOSName string     `json:"vendor_os_name"`
	Qimei        string     `json:"qimei,omitempty"`
}

// 生成随机IMEI号码
func deviceGetRandomIMEI() string {
	imei := make([]int, 14)
	sum := 0

	for i := range imei {
		num := rng.Intn(10)
		if (i+2)%2 == 0 {
			num *= 2
			if num >= 10 {
				num = (num % 10) + 1
			}
		}
		sum += num
		imei[i] = num
	}

	ctrlDigit := (sum * 9) % 10
	return fmt.Sprintf("%s%d", intSliceToString(imei), ctrlDigit)
}

func deviceGetRandomHex(length int) string {
	bytes := make([]byte, length)
	_, _ = rng.Read(bytes)
	return hex.EncodeToString(bytes)
}

func intSliceToString(slice []int) string {
	var builder strings.Builder
	for _, num := range slice {
		builder.WriteString(fmt.Sprintf("%d", num))
	}
	return builder.String()
}

func deviceGetRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rng.Intn(len(charset))]
	}
	return string(result)
}

// NewDevice 创建新的设备信息
func NewDevice() *Device {
	imsiMD5 := make([]byte, 16)
	rng.Read(imsiMD5)
	imsiMD5Arr := md5.Sum(imsiMD5)

	fingerprint := fmt.Sprintf(
		"xiaomi/iarim/sagit:10/eomam.200122.001/%d:user/release-keys",
		rng.Intn(9000000)+1000000,
	)

	device := &Device{
		Display:     fmt.Sprintf("QMAPI.%d.001", rng.Intn(900000)+100000),
		Product:     "iarim",
		Device:      "sagit",
		Board:       "eomam",
		Model:       "MI 6",
		Fingerprint: fingerprint,
		BootID:      uuid.New().String(),
		ProcVersion: fmt.Sprintf(
			"Linux 5.4.0-54-generic-%s (android-build@google.com)",
			deviceGetRandomString(8),
		),
		IMEI:         deviceGetRandomIMEI(),
		Brand:        "Xiaomi",
		Bootloader:   "U-boot",
		BaseBand:     "",
		Version:      newOSVersion(),
		SimInfo:      "T-Mobile",
		OsType:       "android",
		MacAddress:   "00:50:56:C0:00:08",
		WifiBSSID:    "00:50:56:C0:00:08",
		WifiSSID:     "<unknown ssid>",
		IMSIMD5:      imsiMD5Arr[:],
		AndroidID:    deviceGetRandomHex(8),
		APN:          "wifi",
		VendorName:   "MIUI",
		VendorOSName: "qmapi",
		Qimei:        "",
	}
	return device
}
