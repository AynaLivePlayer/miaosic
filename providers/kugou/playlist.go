package kugou

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/aynakeya/deepcolor"
	"github.com/tidwall/gjson"
	"net/http"
	"regexp"
	"strings"
)

var playlistIdRegex = regexp.MustCompile(`gcid_(\w+)`)

// collection_3_600319512_2_0
// collection_3_806499027_106_0
var playlistIdRegex2 = regexp.MustCompile(`collection_\d+_\d+_\d+_\d+`)

func (k *Kugou) MatchPlaylist(uri string) (miaosic.MetaData, bool) {
	if playlistIdRegex.MatchString(uri) {
		matches := playlistIdRegex.FindStringSubmatch(uri)
		return miaosic.MetaData{
			Provider:   k.GetName(),
			Identifier: "gcid_" + matches[1],
		}, true
	}
	if playlistIdRegex2.MatchString(uri) {
		matches := playlistIdRegex2.FindStringSubmatch(uri)
		return miaosic.MetaData{
			Provider:   k.GetName(),
			Identifier: matches[0],
		}, true
	}
	return miaosic.MetaData{}, false
}

func (k *Kugou) getCollectionId(identifier string) (string, error) {
	if strings.HasPrefix(identifier, "collection_") {
		return identifier, nil
	}
	data := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id": identifier, "id_type": "1",
			},
		},
		"ret_info": 1,
	}
	dataBytes, _ := json.Marshal(data)
	param := k.addAndroidParams(map[string]interface{}{}, string(dataBytes))
	urlReq, _ := deepcolor.NewGetRequestWithQuery(
		"https://t.kugou.com/v1/songlist/batch_decode",
		param, map[string]string{},
	)
	urlReq.Method = http.MethodPost
	urlReq.Data = dataBytes
	resp, err := miaosic.Requester.HTTP(urlReq)
	if err != nil {
		return "", err
	}
	collId := gjson.Get(resp.String(), "data.list.0.global_collection_id").String()
	if collId == "" {
		return "", fmt.Errorf("kugou: failed to get collection id")
	}
	return collId, nil
}

func (k *Kugou) getPlaylistTitle(collId string) (string, error) {
	data := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"global_collection_id": collId,
			},
		},
		"userid": "0",
		"token":  "",
	}
	dataBytes, _ := json.Marshal(data)
	param := k.addAndroidParams(map[string]interface{}{}, string(dataBytes))
	urlReq, _ := deepcolor.NewGetRequestWithQuery(
		"https://gateway.kugou.com/v3/get_list_info",
		param, map[string]string{
			"x-router": "pubsongs.kugou.com",
		},
	)
	urlReq.Method = http.MethodPost
	urlReq.Data = dataBytes
	resp, err := miaosic.Requester.HTTP(urlReq)
	if err != nil {
		return "", err
	}
	title := gjson.Get(resp.String(), "data.0.name").String()
	if title == "" {
		return "", fmt.Errorf("kugou: failed to get playlist title")
	}
	return title, nil
}

func (k *Kugou) GetPlaylist(meta miaosic.MetaData) (*miaosic.Playlist, error) {
	collId := meta.Identifier
	if !strings.HasPrefix(collId, "collection_") {
		var err error
		collId, err = k.getCollectionId(collId)
		if err != nil {
			return nil, err
		}
	}
	params := map[string]interface{}{
		"global_collection_id": collId,
		"pagesize":             100,
		"plat":                 1,
		"type":                 1,
		"mode":                 1,
		"area_code":            1,
		"begin_idx":            0,
	}
	playlist := &miaosic.Playlist{
		Meta:   meta,
		Title:  "Kugou Collection " + collId,
		Medias: make([]miaosic.MediaInfo, 0),
	}
	title, err := k.getPlaylistTitle(collId)
	if err == nil {
		playlist.Title = title
	}
	for page := 0; page < 30; page++ {
		params["begin_idx"] = page * 100
		urlReq, _ := deepcolor.NewGetRequestWithQuery(
			"https://gateway.kugou.com/pubsongs/v2/get_other_list_file_nofilt",
			k.addAndroidParams(params, ""), map[string]string{},
		)
		resp, err := miaosic.Requester.HTTP(urlReq)
		if err != nil {
			return nil, err
		}
		result := gjson.ParseBytes(resp.Body())
		if result.Get("error_code").Int() != 0 {
			return nil, errors.New("kugou: get playlist error")
		}
		count := int(result.Get("data.count").Int())
		medias := result.Get("data.songs")
		medias.ForEach(func(key, value gjson.Result) bool {
			playlist.Medias = append(playlist.Medias, miaosic.MediaInfo{
				Title:  value.Get("name").String(),
				Cover:  miaosic.Picture{Url: strings.Replace(value.Get("cover").String(), "{size}", "128", 1)},
				Artist: value.Get("singerinfo.0.name").String(),
				Album:  value.Get("albuminfo.name").String(),
				Meta: miaosic.MetaData{
					Provider:   k.GetName(),
					Identifier: value.Get("hash").String(),
				},
			})
			return true
		})
		if page*100+100 >= count {
			break
		}
	}
	return playlist, nil
}
