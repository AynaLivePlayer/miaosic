package qq

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
)

// 签名算法实现
func qqSignMap(request map[string]interface{}) string {
	data, _ := json.Marshal(request)
	return qqSignStr(string(data))
}

func qqSignStr(data string) string {
	hash := md5.Sum([]byte(data))
	md5Str := strings.ToUpper(hex.EncodeToString(hash[:]))

	// 头部处理
	headPos := []int{21, 4, 9, 26, 16, 20, 27, 30}
	head := make([]byte, len(headPos))
	for i, pos := range headPos {
		head[i] = md5Str[pos]
	}

	// 尾部处理
	tailPos := []int{18, 11, 3, 2, 1, 7, 6, 25}
	tail := make([]byte, len(tailPos))
	for i, pos := range tailPos {
		tail[i] = md5Str[pos]
	}

	// 中间部分处理
	ol := []byte{212, 45, 80, 68, 195, 163, 163, 203, 157, 220, 254, 91, 204, 79, 104, 6}
	hexMap := map[byte]byte{
		'0': 0, '1': 1, '2': 2, '3': 3, '4': 4,
		'5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
		'A': 10, 'B': 11, 'C': 12, 'D': 13, 'E': 14, 'F': 15,
	}

	middle := make([]byte, 16)
	for i := 0; i < 32; i += 2 {
		idx := i / 2
		one := hexMap[md5Str[i]]
		two := hexMap[md5Str[i+1]]
		r := one*16 ^ two
		middle[idx] = r ^ ol[idx]
	}

	// 组合签名
	middleBase64 := base64.StdEncoding.EncodeToString(middle)
	signature := "zzb" + string(head) + middleBase64 + string(tail)
	signature = strings.ToLower(signature)
	signature = strings.ReplaceAll(signature, "/", "")
	signature = strings.ReplaceAll(signature, "+", "")
	signature = strings.ReplaceAll(signature, "=", "")
	return signature
}
