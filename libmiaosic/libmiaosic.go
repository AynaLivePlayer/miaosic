package main

/*
#include <stdlib.h>
#include <string.h>
#include "libmiaosic.h"
*/
import "C"

import (
	"errors"
	"unsafe"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/AynaLivePlayer/miaosic/providers/bilivideo"
	"github.com/AynaLivePlayer/miaosic/providers/kugou"
	"github.com/AynaLivePlayer/miaosic/providers/kuwo"
	"github.com/AynaLivePlayer/miaosic/providers/local"
	"github.com/AynaLivePlayer/miaosic/providers/netease"
	"github.com/AynaLivePlayer/miaosic/providers/qq"
)

func init() {
	miaosic.UnregisterAllProvider()
}

func goString(val *C.char) string {
	if val == nil {
		return ""
	}
	return C.GoString(val)
}

func newResult(data unsafe.Pointer, resultType C.MiaosicResultType, err error) *C.MiaosicResult {
	result := (*C.MiaosicResult)(C.malloc(C.size_t(unsafe.Sizeof(C.MiaosicResult{}))))
	if result == nil {
		return nil
	}
	result.ok = 0
	result.err = nil
	result.result_type = resultType
	result.data = data
	if err == nil {
		result.ok = 1
		return result
	}
	result.err = C.CString(err.Error())
	return result
}

func newBoolResult(val bool) unsafe.Pointer {
	ptr := (*C.MiaosicBool)(C.malloc(C.size_t(unsafe.Sizeof(C.MiaosicBool{}))))
	if ptr == nil {
		return nil
	}
	ptr.value = 0
	if val {
		ptr.value = 1
	}
	return unsafe.Pointer(ptr)
}

func newStringResult(val string) unsafe.Pointer {
	ptr := (*C.MiaosicString)(C.malloc(C.size_t(unsafe.Sizeof(C.MiaosicString{}))))
	if ptr == nil {
		return nil
	}
	ptr.value = C.CString(val)
	return unsafe.Pointer(ptr)
}

func fillMetaData(dst *C.MiaosicMetaData, meta miaosic.MetaData) {
	dst.provider = C.CString(meta.Provider)
	dst.identifier = C.CString(meta.Identifier)
}

func fillPicture(dst *C.MiaosicPicture, pic miaosic.Picture) {
	dst.url = nil
	dst.data = nil
	dst.data_len = 0
	if pic.Url != "" {
		dst.url = C.CString(pic.Url)
	}
	if len(pic.Data) > 0 {
		dst.data = (*C.uchar)(C.malloc(C.size_t(len(pic.Data))))
		if dst.data != nil {
			C.memcpy(unsafe.Pointer(dst.data), unsafe.Pointer(&pic.Data[0]), C.size_t(len(pic.Data)))
			dst.data_len = C.int(len(pic.Data))
		}
	}
}

func fillMediaInfo(dst *C.MiaosicMediaInfo, info miaosic.MediaInfo) {
	dst.title = C.CString(info.Title)
	dst.artist = C.CString(info.Artist)
	dst.album = C.CString(info.Album)
	fillPicture(&dst.cover, info.Cover)
	fillMetaData(&dst.meta, info.Meta)
}

func newMediaInfoResult(info miaosic.MediaInfo) unsafe.Pointer {
	ptr := (*C.MiaosicMediaInfo)(C.malloc(C.size_t(unsafe.Sizeof(C.MiaosicMediaInfo{}))))
	if ptr == nil {
		return nil
	}
	fillMediaInfo(ptr, info)
	return unsafe.Pointer(ptr)
}

func newMediaInfoListResult(items []miaosic.MediaInfo) unsafe.Pointer {
	list := (*C.MiaosicMediaInfoList)(C.malloc(C.size_t(unsafe.Sizeof(C.MiaosicMediaInfoList{}))))
	if list == nil {
		return nil
	}
	list.len = C.int(len(items))
	list.items = nil
	if len(items) == 0 {
		return unsafe.Pointer(list)
	}
	list.items = (*C.MiaosicMediaInfo)(C.malloc(C.size_t(len(items)) * C.size_t(unsafe.Sizeof(C.MiaosicMediaInfo{}))))
	if list.items == nil {
		list.len = 0
		return unsafe.Pointer(list)
	}
	slice := (*[1 << 30]C.MiaosicMediaInfo)(unsafe.Pointer(list.items))[:len(items):len(items)]
	for i, item := range items {
		fillMediaInfo(&slice[i], item)
	}
	return unsafe.Pointer(list)
}

func fillHeaderPairs(headers map[string]string) (*C.MiaosicHeaderPair, C.int) {
	if len(headers) == 0 {
		return nil, 0
	}
	ptr := (*C.MiaosicHeaderPair)(C.malloc(C.size_t(len(headers)) * C.size_t(unsafe.Sizeof(C.MiaosicHeaderPair{}))))
	if ptr == nil {
		return nil, 0
	}
	slice := (*[1 << 30]C.MiaosicHeaderPair)(unsafe.Pointer(ptr))[:len(headers):len(headers)]
	idx := 0
	for key, value := range headers {
		slice[idx].key = C.CString(key)
		slice[idx].value = C.CString(value)
		idx++
	}
	return ptr, C.int(len(headers))
}

func fillMediaUrl(dst *C.MiaosicMediaUrl, url miaosic.MediaUrl) {
	dst.url = C.CString(url.Url)
	dst.quality = C.CString(string(url.Quality))
	dst.headers = nil
	dst.header_len = 0
	headers, headerLen := fillHeaderPairs(url.Header)
	dst.headers = headers
	dst.header_len = headerLen
}

func newMediaUrlListResult(items []miaosic.MediaUrl) unsafe.Pointer {
	list := (*C.MiaosicMediaUrlList)(C.malloc(C.size_t(unsafe.Sizeof(C.MiaosicMediaUrlList{}))))
	if list == nil {
		return nil
	}
	list.len = C.int(len(items))
	list.items = nil
	if len(items) == 0 {
		return unsafe.Pointer(list)
	}
	list.items = (*C.MiaosicMediaUrl)(C.malloc(C.size_t(len(items)) * C.size_t(unsafe.Sizeof(C.MiaosicMediaUrl{}))))
	if list.items == nil {
		list.len = 0
		return unsafe.Pointer(list)
	}
	slice := (*[1 << 30]C.MiaosicMediaUrl)(unsafe.Pointer(list.items))[:len(items):len(items)]
	for i, item := range items {
		fillMediaUrl(&slice[i], item)
	}
	return unsafe.Pointer(list)
}

func newPlaylistResult(playlist *miaosic.Playlist) unsafe.Pointer {
	if playlist == nil {
		return nil
	}
	ptr := (*C.MiaosicPlaylist)(C.malloc(C.size_t(unsafe.Sizeof(C.MiaosicPlaylist{}))))
	if ptr == nil {
		return nil
	}
	ptr.title = C.CString(playlist.Title)
	fillMetaData(&ptr.meta, playlist.Meta)
	medias := make([]miaosic.MediaInfo, len(playlist.Medias))
	copy(medias, playlist.Medias)
	listPtr := newMediaInfoListResult(medias)
	if listPtr != nil {
		ptr.medias = *(*C.MiaosicMediaInfoList)(listPtr)
		C.free(listPtr)
	} else {
		ptr.medias.len = 0
		ptr.medias.items = nil
	}
	return unsafe.Pointer(ptr)
}

func newMatchResult(meta miaosic.MetaData, matched bool) unsafe.Pointer {
	ptr := (*C.MiaosicMatchResult)(C.malloc(C.size_t(unsafe.Sizeof(C.MiaosicMatchResult{}))))
	if ptr == nil {
		return nil
	}
	ptr.matched = 0
	if matched {
		ptr.matched = 1
	}
	fillMetaData(&ptr.meta, meta)
	return unsafe.Pointer(ptr)
}

func newQrLoginSessionResult(session *miaosic.QrLoginSession) unsafe.Pointer {
	if session == nil {
		return nil
	}
	ptr := (*C.MiaosicQrLoginSession)(C.malloc(C.size_t(unsafe.Sizeof(C.MiaosicQrLoginSession{}))))
	if ptr == nil {
		return nil
	}
	ptr.url = C.CString(session.Url)
	ptr.key = C.CString(session.Key)
	return unsafe.Pointer(ptr)
}

func newQrLoginResult(res *miaosic.QrLoginResult) unsafe.Pointer {
	if res == nil {
		return nil
	}
	ptr := (*C.MiaosicQrLoginResult)(C.malloc(C.size_t(unsafe.Sizeof(C.MiaosicQrLoginResult{}))))
	if ptr == nil {
		return nil
	}
	ptr.success = 0
	if res.Success {
		ptr.success = 1
	}
	ptr.message = C.CString(res.Message)
	return unsafe.Pointer(ptr)
}

func newLyricsListResult(items []miaosic.Lyrics) unsafe.Pointer {
	list := (*C.MiaosicLyricsList)(C.malloc(C.size_t(unsafe.Sizeof(C.MiaosicLyricsList{}))))
	if list == nil {
		return nil
	}
	list.len = C.int(len(items))
	list.items = nil
	if len(items) == 0 {
		return unsafe.Pointer(list)
	}
	list.items = (*C.MiaosicLyrics)(C.malloc(C.size_t(len(items)) * C.size_t(unsafe.Sizeof(C.MiaosicLyrics{}))))
	if list.items == nil {
		list.len = 0
		return unsafe.Pointer(list)
	}
	slice := (*[1 << 30]C.MiaosicLyrics)(unsafe.Pointer(list.items))[:len(items):len(items)]
	for i, item := range items {
		slice[i].lang = C.CString(item.Lang)
		slice[i].lyrics = C.CString(item.String())
	}
	return unsafe.Pointer(list)
}

func newStringListResult(items []string) unsafe.Pointer {
	list := (*C.MiaosicStringList)(C.malloc(C.size_t(unsafe.Sizeof(C.MiaosicStringList{}))))
	if list == nil {
		return nil
	}
	list.len = C.int(len(items))
	list.items = nil
	if len(items) == 0 {
		return unsafe.Pointer(list)
	}
	list.items = (**C.char)(C.malloc(C.size_t(len(items)) * C.size_t(unsafe.Sizeof((*C.char)(nil)))))
	if list.items == nil {
		list.len = 0
		return unsafe.Pointer(list)
	}
	slice := (*[1 << 30]*C.char)(unsafe.Pointer(list.items))[:len(items):len(items)]
	for i, item := range items {
		slice[i] = C.CString(item)
	}
	return unsafe.Pointer(list)
}

func registerProviderIfNeeded(provider miaosic.MediaProvider) error {
	if provider == nil {
		return errors.New("miaosic: provider is nil")
	}
	if _, ok := miaosic.GetProvider(provider.GetName()); ok {
		return nil
	}
	miaosic.RegisterProvider(provider)
	return nil
}

func freeMetaData(meta *C.MiaosicMetaData) {
	if meta == nil {
		return
	}
	if meta.provider != nil {
		C.free(unsafe.Pointer(meta.provider))
	}
	if meta.identifier != nil {
		C.free(unsafe.Pointer(meta.identifier))
	}
}

func freePicture(pic *C.MiaosicPicture) {
	if pic == nil {
		return
	}
	if pic.url != nil {
		C.free(unsafe.Pointer(pic.url))
	}
	if pic.data != nil {
		C.free(unsafe.Pointer(pic.data))
	}
}

func freeMediaInfo(info *C.MiaosicMediaInfo) {
	if info == nil {
		return
	}
	if info.title != nil {
		C.free(unsafe.Pointer(info.title))
	}
	if info.artist != nil {
		C.free(unsafe.Pointer(info.artist))
	}
	if info.album != nil {
		C.free(unsafe.Pointer(info.album))
	}
	freePicture(&info.cover)
	freeMetaData(&info.meta)
}

func freeHeaderPairs(headers *C.MiaosicHeaderPair, length C.int) {
	if headers == nil || length == 0 {
		return
	}
	slice := (*[1 << 30]C.MiaosicHeaderPair)(unsafe.Pointer(headers))[:length:length]
	for i := 0; i < int(length); i++ {
		if slice[i].key != nil {
			C.free(unsafe.Pointer(slice[i].key))
		}
		if slice[i].value != nil {
			C.free(unsafe.Pointer(slice[i].value))
		}
	}
	C.free(unsafe.Pointer(headers))
}

func freeMediaUrl(url *C.MiaosicMediaUrl) {
	if url == nil {
		return
	}
	if url.url != nil {
		C.free(unsafe.Pointer(url.url))
	}
	if url.quality != nil {
		C.free(unsafe.Pointer(url.quality))
	}
	freeHeaderPairs(url.headers, url.header_len)
}

func freeMediaInfoList(list *C.MiaosicMediaInfoList) {
	if list == nil {
		return
	}
	if list.items != nil && list.len > 0 {
		slice := (*[1 << 30]C.MiaosicMediaInfo)(unsafe.Pointer(list.items))[:list.len:list.len]
		for i := 0; i < int(list.len); i++ {
			freeMediaInfo(&slice[i])
		}
		C.free(unsafe.Pointer(list.items))
	}
}

func freeMediaUrlList(list *C.MiaosicMediaUrlList) {
	if list == nil {
		return
	}
	if list.items != nil && list.len > 0 {
		slice := (*[1 << 30]C.MiaosicMediaUrl)(unsafe.Pointer(list.items))[:list.len:list.len]
		for i := 0; i < int(list.len); i++ {
			freeMediaUrl(&slice[i])
		}
		C.free(unsafe.Pointer(list.items))
	}
}

func freePlaylist(playlist *C.MiaosicPlaylist) {
	if playlist == nil {
		return
	}
	if playlist.title != nil {
		C.free(unsafe.Pointer(playlist.title))
	}
	freeMetaData(&playlist.meta)
	freeMediaInfoList(&playlist.medias)
}

func freeMatchResult(result *C.MiaosicMatchResult) {
	if result == nil {
		return
	}
	freeMetaData(&result.meta)
}

func freeQrLoginSession(session *C.MiaosicQrLoginSession) {
	if session == nil {
		return
	}
	if session.url != nil {
		C.free(unsafe.Pointer(session.url))
	}
	if session.key != nil {
		C.free(unsafe.Pointer(session.key))
	}
}

func freeQrLoginResult(result *C.MiaosicQrLoginResult) {
	if result == nil {
		return
	}
	if result.message != nil {
		C.free(unsafe.Pointer(result.message))
	}
}

func freeLyricsList(list *C.MiaosicLyricsList) {
	if list == nil {
		return
	}
	if list.items != nil && list.len > 0 {
		slice := (*[1 << 30]C.MiaosicLyrics)(unsafe.Pointer(list.items))[:list.len:list.len]
		for i := 0; i < int(list.len); i++ {
			if slice[i].lang != nil {
				C.free(unsafe.Pointer(slice[i].lang))
			}
			if slice[i].lyrics != nil {
				C.free(unsafe.Pointer(slice[i].lyrics))
			}
		}
		C.free(unsafe.Pointer(list.items))
	}
}

func freeStringList(list *C.MiaosicStringList) {
	if list == nil {
		return
	}
	if list.items != nil && list.len > 0 {
		slice := (*[1 << 30]*C.char)(unsafe.Pointer(list.items))[:list.len:list.len]
		for i := 0; i < int(list.len); i++ {
			if slice[i] != nil {
				C.free(unsafe.Pointer(slice[i]))
			}
		}
		C.free(unsafe.Pointer(list.items))
	}
}

//export FreeResult
func FreeResult(res *C.MiaosicResult) {
	if res == nil {
		return
	}
	if res.err != nil {
		C.free(unsafe.Pointer(res.err))
	}
	if res.data != nil {
		switch res.result_type {
		case C.MIAOSIC_RESULT_BOOL:
			C.free(res.data)
		case C.MIAOSIC_RESULT_STRING:
			str := (*C.MiaosicString)(res.data)
			if str.value != nil {
				C.free(unsafe.Pointer(str.value))
			}
			C.free(res.data)
		case C.MIAOSIC_RESULT_STRING_LIST:
			list := (*C.MiaosicStringList)(res.data)
			freeStringList(list)
			C.free(res.data)
		case C.MIAOSIC_RESULT_META:
			meta := (*C.MiaosicMetaData)(res.data)
			freeMetaData(meta)
			C.free(res.data)
		case C.MIAOSIC_RESULT_MEDIA_INFO:
			info := (*C.MiaosicMediaInfo)(res.data)
			freeMediaInfo(info)
			C.free(res.data)
		case C.MIAOSIC_RESULT_MEDIA_INFO_LIST:
			list := (*C.MiaosicMediaInfoList)(res.data)
			freeMediaInfoList(list)
			C.free(res.data)
		case C.MIAOSIC_RESULT_MEDIA_URL_LIST:
			list := (*C.MiaosicMediaUrlList)(res.data)
			freeMediaUrlList(list)
			C.free(res.data)
		case C.MIAOSIC_RESULT_PLAYLIST:
			playlist := (*C.MiaosicPlaylist)(res.data)
			freePlaylist(playlist)
			C.free(res.data)
		case C.MIAOSIC_RESULT_MATCH:
			match := (*C.MiaosicMatchResult)(res.data)
			freeMatchResult(match)
			C.free(res.data)
		case C.MIAOSIC_RESULT_QR_LOGIN_SESSION:
			session := (*C.MiaosicQrLoginSession)(res.data)
			freeQrLoginSession(session)
			C.free(res.data)
		case C.MIAOSIC_RESULT_QR_LOGIN_RESULT:
			result := (*C.MiaosicQrLoginResult)(res.data)
			freeQrLoginResult(result)
			C.free(res.data)
		case C.MIAOSIC_RESULT_LYRICS_LIST:
			list := (*C.MiaosicLyricsList)(res.data)
			freeLyricsList(list)
			C.free(res.data)
		default:
			C.free(res.data)
		}
	}
	C.free(unsafe.Pointer(res))
}

//export SearchByProvider
func SearchByProvider(provider *C.char, keyword *C.char, page C.int, size C.int) *C.MiaosicResult {
	data, err := miaosic.SearchByProvider(goString(provider), goString(keyword), int(page), int(size))
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_MEDIA_INFO_LIST, err)
	}
	resultData := newMediaInfoListResult(data)
	return newResult(resultData, C.MIAOSIC_RESULT_MEDIA_INFO_LIST, err)
}

//export GetMediaUrl
func GetMediaUrl(provider *C.char, identifier *C.char, quality *C.char) *C.MiaosicResult {
	meta := miaosic.NewMetaData(goString(provider), goString(identifier))
	data, err := miaosic.GetMediaUrl(meta, miaosic.Quality(goString(quality)))
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_MEDIA_URL_LIST, err)
	}
	resultData := newMediaUrlListResult(data)
	return newResult(resultData, C.MIAOSIC_RESULT_MEDIA_URL_LIST, err)
}

//export GetMediaInfo
func GetMediaInfo(provider *C.char, identifier *C.char) *C.MiaosicResult {
	meta := miaosic.NewMetaData(goString(provider), goString(identifier))
	data, err := miaosic.GetMediaInfo(meta)
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_MEDIA_INFO, err)
	}
	resultData := newMediaInfoResult(data)
	return newResult(resultData, C.MIAOSIC_RESULT_MEDIA_INFO, err)
}

//export GetMediaLyric
func GetMediaLyric(provider *C.char, identifier *C.char) *C.MiaosicResult {
	meta := miaosic.NewMetaData(goString(provider), goString(identifier))
	data, err := miaosic.GetMediaLyric(meta)
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_LYRICS_LIST, err)
	}
	resultData := newLyricsListResult(data)
	return newResult(resultData, C.MIAOSIC_RESULT_LYRICS_LIST, err)
}

//export MatchPlaylistByProvider
func MatchPlaylistByProvider(provider *C.char, uri *C.char) *C.MiaosicResult {
	meta, matched := miaosic.MatchPlaylistByProvider(goString(provider), goString(uri))
	resultData := newMatchResult(meta, matched)
	return newResult(resultData, C.MIAOSIC_RESULT_MATCH, nil)
}

//export GetPlaylist
func GetPlaylist(provider *C.char, identifier *C.char) *C.MiaosicResult {
	meta := miaosic.NewMetaData(goString(provider), goString(identifier))
	data, err := miaosic.GetPlaylist(meta)
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_PLAYLIST, err)
	}
	resultData := newPlaylistResult(data)
	return newResult(resultData, C.MIAOSIC_RESULT_PLAYLIST, err)
}

//export MatchMedia
func MatchMedia(keyword *C.char) *C.MiaosicResult {
	meta, matched := miaosic.MatchMedia(goString(keyword))
	resultData := newMatchResult(meta, matched)
	return newResult(resultData, C.MIAOSIC_RESULT_MATCH, nil)
}

//export MatchMediaByProvider
func MatchMediaByProvider(provider *C.char, uri *C.char) *C.MiaosicResult {
	meta, matched := miaosic.MatchMediaByProvider(goString(provider), goString(uri))
	resultData := newMatchResult(meta, matched)
	return newResult(resultData, C.MIAOSIC_RESULT_MATCH, nil)
}

//export ListAvailableProviders
func ListAvailableProviders() *C.MiaosicResult {
	data := miaosic.ListAvailableProviders()
	resultData := newStringListResult(data)
	return newResult(resultData, C.MIAOSIC_RESULT_STRING_LIST, nil)
}

//export LoginByProvider
func LoginByProvider(provider *C.char, username *C.char, password *C.char) *C.MiaosicResult {
	err := miaosic.LoginByProvider(goString(provider), goString(username), goString(password))
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_BOOL, err)
	}
	resultData := newBoolResult(true)
	return newResult(resultData, C.MIAOSIC_RESULT_BOOL, nil)
}

//export LogoutByProvider
func LogoutByProvider(provider *C.char) *C.MiaosicResult {
	err := miaosic.LogoutByProvider(goString(provider))
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_BOOL, err)
	}
	resultData := newBoolResult(true)
	return newResult(resultData, C.MIAOSIC_RESULT_BOOL, nil)
}

//export IsLoginByProvider
func IsLoginByProvider(provider *C.char) *C.MiaosicResult {
	data, err := miaosic.IsLoginByProvider(goString(provider))
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_BOOL, err)
	}
	resultData := newBoolResult(data)
	return newResult(resultData, C.MIAOSIC_RESULT_BOOL, nil)
}

//export RefreshLoginByProvider
func RefreshLoginByProvider(provider *C.char) *C.MiaosicResult {
	err := miaosic.RefreshLoginByProvider(goString(provider))
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_BOOL, err)
	}
	resultData := newBoolResult(true)
	return newResult(resultData, C.MIAOSIC_RESULT_BOOL, nil)
}

//export QrLoginByProvider
func QrLoginByProvider(provider *C.char) *C.MiaosicResult {
	data, err := miaosic.QrLoginByProvider(goString(provider))
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_QR_LOGIN_SESSION, err)
	}
	resultData := newQrLoginSessionResult(data)
	return newResult(resultData, C.MIAOSIC_RESULT_QR_LOGIN_SESSION, nil)
}

//export QrLoginVerifyByProvider
func QrLoginVerifyByProvider(provider *C.char, key *C.char, url *C.char) *C.MiaosicResult {
	session := &miaosic.QrLoginSession{
		Key: goString(key),
		Url: goString(url),
	}
	data, err := miaosic.QrLoginVerifyByProvider(goString(provider), session)
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_QR_LOGIN_RESULT, err)
	}
	resultData := newQrLoginResult(data)
	return newResult(resultData, C.MIAOSIC_RESULT_QR_LOGIN_RESULT, nil)
}

//export RestoreSessionByProvider
func RestoreSessionByProvider(provider *C.char, session *C.char) *C.MiaosicResult {
	err := miaosic.RestoreSessionByProvider(goString(provider), goString(session))
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_BOOL, err)
	}
	resultData := newBoolResult(true)
	return newResult(resultData, C.MIAOSIC_RESULT_BOOL, nil)
}

//export SaveSessionByProvider
func SaveSessionByProvider(provider *C.char) *C.MiaosicResult {
	data, err := miaosic.SaveSessionByProvider(goString(provider))
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_STRING, err)
	}
	resultData := newStringResult(data)
	return newResult(resultData, C.MIAOSIC_RESULT_STRING, nil)
}

//export UseBilibiliVideo
func UseBilibiliVideo() *C.MiaosicResult {
	err := registerProviderIfNeeded(bilivideo.NewBilibiliViedo())
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_BOOL, err)
	}
	resultData := newBoolResult(true)
	return newResult(resultData, C.MIAOSIC_RESULT_BOOL, nil)
}

//export UseKugou
func UseKugou() *C.MiaosicResult {
	err := registerProviderIfNeeded(kugou.NewKugou(false))
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_BOOL, err)
	}
	resultData := newBoolResult(true)
	return newResult(resultData, C.MIAOSIC_RESULT_BOOL, nil)
}

//export UseKugouInstrumental
func UseKugouInstrumental() *C.MiaosicResult {
	if _, ok := miaosic.GetProvider("kugou-instr"); ok {
		return newResult(newBoolResult(true), C.MIAOSIC_RESULT_BOOL, nil)
	}
	kugou.UseInstrumental()
	return newResult(newBoolResult(true), C.MIAOSIC_RESULT_BOOL, nil)
}

//export UseKuwo
func UseKuwo() *C.MiaosicResult {
	err := registerProviderIfNeeded(kuwo.NewKuwo())
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_BOOL, err)
	}
	resultData := newBoolResult(true)
	return newResult(resultData, C.MIAOSIC_RESULT_BOOL, nil)
}

//export UseNetease
func UseNetease() *C.MiaosicResult {
	err := registerProviderIfNeeded(netease.NewNetease())
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_BOOL, err)
	}
	resultData := newBoolResult(true)
	return newResult(resultData, C.MIAOSIC_RESULT_BOOL, nil)
}

//export UseQQLogin
func UseQQLogin() *C.MiaosicResult {
	if _, ok := miaosic.GetProvider("qq"); ok {
		return newResult(newBoolResult(true), C.MIAOSIC_RESULT_BOOL, nil)
	}
	qq.UseQQLogin()
	return newResult(newBoolResult(true), C.MIAOSIC_RESULT_BOOL, nil)
}

//export UseWechatLogin
func UseWechatLogin() *C.MiaosicResult {
	if _, ok := miaosic.GetProvider("qq"); ok {
		return newResult(newBoolResult(true), C.MIAOSIC_RESULT_BOOL, nil)
	}
	qq.UseWechatLogin()
	return newResult(newBoolResult(true), C.MIAOSIC_RESULT_BOOL, nil)
}

//export UseLocal
func UseLocal(localDir *C.char) *C.MiaosicResult {
	err := registerProviderIfNeeded(local.NewLocal(goString(localDir)))
	if err != nil {
		return newResult(nil, C.MIAOSIC_RESULT_BOOL, err)
	}
	resultData := newBoolResult(true)
	return newResult(resultData, C.MIAOSIC_RESULT_BOOL, nil)
}

func main() {}
