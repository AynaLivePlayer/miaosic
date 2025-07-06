package qq

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/AynaLivePlayer/miaosic/providers/qq/goqrcdec"
	"math/rand"
	"strconv"
	"time"
)

var rng *rand.Rand

func calcMd5(data ...string) string {
	h := md5.New()
	for _, d := range data {
		_, _ = h.Write([]byte(d))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func rsaEncrypt(plainText []byte, publicKeyPEM string) ([]byte, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not an RSA public key")
	}

	return rsa.EncryptPKCS1v15(crand.Reader, rsaPub, plainText)
}

func aesEncrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 填充数据
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	plaintext = append(plaintext, padText...)

	// 使用key作为IV
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := key[:aes.BlockSize]
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext[aes.BlockSize:], nil
}

func getSearchID() string {
	/* 随机 searchID

	Returns:
		随机 searchID
	*/
	e := rng.Intn(20) + 1
	t := e * 18014398509481984
	n := rng.Intn(4194305) * 4294967296
	a := time.Now().UnixNano() / int64(time.Millisecond)
	r := a % (24 * 60 * 60 * 1000)
	return strconv.FormatInt(int64(t)+int64(n)+r, 10)
}

func getGuid() string {
	const charset = "abcdef1234567890"
	result := make([]byte, 32)
	for i := range result {
		result[i] = charset[rng.Intn(len(charset))]
	}
	return string(result)
}

// 计算hash33
func hash33(s string, h int) int {
	val := uint64(h)
	for _, c := range []byte(s) {
		val = (val << 5) + val + uint64(c)
	}
	return int(2147483647 & val)
}

func qrcDecrypt(hexStr string) (string, error) {
	// 1. hex 解码
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}

	val, err := goqrcdec.DecodeQRC(data)

	return string(val), err
}
