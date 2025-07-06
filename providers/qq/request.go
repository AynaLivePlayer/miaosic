package qq

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/aynakeya/deepcolor/dphttp"
	"github.com/tidwall/gjson"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (p *QQMusicProvider) makeApiRequest(module, method string, params map[string]interface{}) (gjson.Result, error) {
	if !p.qimeiUpdated {
		_, _ = getQimei(p.device, p.cfg.Version)
		p.qimeiUpdated = true
	}

	expiredTime := time.UnixMilli(p.cred.ExpiredAt * 1000)

	if expiredTime.Before(time.Now().Add(24*time.Hour)) && !p.tokenRefreshed {
		_ = p.refreshToken()
		// only refresh once
		p.tokenRefreshed = true
	}

	// 公共参数
	common := map[string]interface{}{
		"ct":         "11",
		"tmeAppID":   "qqmusic",
		"format":     "json",
		"inCharset":  "utf-8",
		"outCharset": "utf-8",
		"uid":        "3931641530",
		"cv":         p.cfg.VersionCode,
		"v":          p.cfg.VersionCode,
		"QIMEI36":    p.device.Qimei,
	}

	cookie := map[string]interface{}{}

	if p.cred.HasMusicKey() && p.cred.HasMusicID() {
		common["authst"] = p.cred.MusicKey
		common["qq"] = p.cred.MusicID
		common["tmeLoginType"] = strconv.Itoa(p.cred.GetFormatedLoginType())
		cookie["uin"] = p.cred.MusicID
		cookie["qqmusic_key"] = p.cred.MusicKey
		cookie["qm_keyst"] = p.cred.MusicKey
		cookie["tmeLoginType"] = strconv.Itoa(p.cred.GetFormatedLoginType())
	}

	moduleKey := fmt.Sprintf("%s.%s", module, method)

	requestData := map[string]interface{}{
		"comm": common,
		moduleKey: map[string]interface{}{
			"module": module,
			"method": method,
			"param":  params,
		},
	}
	jsonData, _ := json.Marshal(requestData)

	uri := p.cfg.Endpoint
	if p.cfg.EnableSign {
		// 创建请求
		uri = p.cfg.EncEndpoint + "?sign=" + url.QueryEscape(qqSignStr(string(jsonData)))
	}

	request := dphttp.Request{
		Method: http.MethodPost,
		Url:    dphttp.UrlMustParse(uri),
		Header: map[string]string{
			"Referer":      "https://y.qq.com/",
			"Content-Type": "application/json",
			"User-Agent":   "Mozilla/5.0 (Windows NT 11.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36 Edg/116.0.1938.54",
		},
		Data: jsonData,
	}

	cookieStr := ""
	for k, v := range cookie {
		cookieStr += fmt.Sprintf("%s=%s;", k, v)
	}
	if cookieStr != "" {
		request.Header["Cookie"] = cookieStr
	}

	response, err := miaosic.Requester.HTTP(&request)
	if err != nil {
		return gjson.Result{}, err
	}
	jsonResp := gjson.ParseBytes(response.Body())
	//fmt.Println(response.String())
	moduleKeyEscaped := strings.ReplaceAll(moduleKey, ".", "\\.")
	if !jsonResp.Get(moduleKeyEscaped).Exists() {
		return gjson.Result{}, fmt.Errorf("miaosic (qq): api request fail")
	}
	code := jsonResp.Get(moduleKeyEscaped + ".code").Int()
	if code == 4000 {
		return jsonResp.Get(moduleKeyEscaped), errors.New("miaosic (qq): not login")
	}
	if code == 2000 {
		return jsonResp.Get(moduleKeyEscaped), errors.New("miaosic (qq): invalid signature")
	}
	if code == 1000 {
		return jsonResp.Get(moduleKeyEscaped), errors.New("miaosic (qq): invalid cookie")
	}
	if code != 0 {
		return jsonResp.Get(moduleKeyEscaped), fmt.Errorf("miaosic (qq): invalid code: %d", code)
	}
	return jsonResp.Get(moduleKeyEscaped), nil
}
