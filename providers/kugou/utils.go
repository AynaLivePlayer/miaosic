package kugou

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	appid         = "1005"
	clientver     = "12329"
	appidLite     = "3116"
	clientverLite = "10940"

	signkey     = "OIlwieks28dk2k092lksi2UIkp"
	signkeyLite = "LnT6xpN3khm36zse0QzvmgTZ3waWdRSA"
)

// signKey encrypts the given parameters and returns the encrypted sign.
func signKey(appid string, hash, mid, userid string) string {
	data := hash + "57ae12eb6890223e355ccfcb74edf70d" + appid + mid + userid
	return getMD5Hash(data)
}

func signatureAndroidParams(signkey string, params map[string]interface{}, data string) string {
	// Sort the keys of the params map
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Create the params string
	var paramsString strings.Builder
	for _, key := range keys {
		value := params[key]
		var valueStr string
		switch v := value.(type) {
		case map[string]interface{}, []interface{}:
			jsonValue, _ := json.Marshal(v)
			valueStr = string(jsonValue)
		default:
			valueStr = fmt.Sprintf("%v", v)
		}
		paramsString.WriteString(fmt.Sprintf("%s=%s", key, valueStr))
	}

	// Generate the MD5 hash
	hash := md5.Sum([]byte(signkey + paramsString.String() + data + signkey))
	return hex.EncodeToString(hash[:])
}

// signatureWebParams generates a signature for the given parameters.
func signatureWebParams(params map[string]string) string {
	str := "NVPh5oo715z5DIWAeQlhMDsWXXQV4hwt"

	// Sort the keys of the params map
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Create the params string
	var paramsString strings.Builder
	for _, key := range keys {
		paramsString.WriteString(fmt.Sprintf("%s=%s", key, params[key]))
	}

	// Generate the MD5 hash
	hash := md5.Sum([]byte(str + paramsString.String() + str))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func (k *Kugou) addAndroidParams(params map[string]interface{}, data string) map[string]interface{} {
	if token, ok := k.cookie["token"]; ok {
		params["token"] = token
	} else {
		params["token"] = ""
	}
	if userId, ok := k.cookie["userid"]; ok {
		params["userid"] = userId
	} else {
		params["userid"] = "0"
	}
	params["appid"] = k.appid
	params["clientver"] = k.clientver
	params["dfid"] = k.dfid
	params["mid"] = getMD5Hash(k.dfid)
	params["uuid"] = getMD5Hash(k.dfid)
	params["clienttime"] = fmt.Sprintf("%d", time.Now().Unix())
	params["signature"] = signatureAndroidParams(k.signkey, params, data)
	return params
}
