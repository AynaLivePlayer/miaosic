package miaosic

func SearchByProvider(provider string, keyword string, page, size int) ([]MediaInfo, error) {
	p, ok := GetProvider(provider)
	if !ok {
		return nil, ErrorNoSuchProvider
	}
	return p.Search(keyword, page, size)
}

func GetMediaUrl(meta MetaData, quality Quality) ([]MediaUrl, error) {
	provider, ok := GetProvider(meta.Provider)
	if !ok {
		return nil, ErrorNoSuchProvider
	}
	return provider.GetMediaUrl(meta, quality)
}

func GetMediaInfo(meta MetaData) (MediaInfo, error) {
	provider, ok := GetProvider(meta.Provider)
	if !ok {
		return MediaInfo{}, ErrorNoSuchProvider
	}
	return provider.GetMediaInfo(meta)
}

func GetMediaLyric(meta MetaData) ([]Lyrics, error) {
	provider, ok := GetProvider(meta.Provider)
	if !ok {
		return nil, ErrorNoSuchProvider
	}
	return provider.GetMediaLyric(meta)
}

//func GetPlaylist(meta *model.Meta) ([]*model.Media, error) {
//	if v, ok := Providers[meta.Name]; ok {
//		return v.GetPlaylist(meta)
//	}
//	return nil, ErrorNoSuchProvider
//}
//
//func FormatPlaylistUrl(pname, uri string) (string, error) {
//	if v, ok := Providers[pname]; ok {
//		return v.FormatPlaylistUrl(uri), nil
//	}
//	return "", ErrorNoSuchProvider
//}
//
//func MatchMedia(provider string, keyword string) *model.Media {
//	if v, ok := Providers[provider]; ok {
//		return v.MatchMedia(keyword)
//	}
//	return nil
//}
//
//func Search(provider string, keyword string) ([]*model.Media, error) {
//	if v, ok := Providers[provider]; ok {
//		return v.Search(keyword)
//	}
//	return nil, ErrorNoSuchProvider
//}
//
//func UpdateMedia(media *model.Media) error {
//	if v, ok := Providers[media.Meta.(model.Meta).Name]; ok {
//		return v.UpdateMedia(media)
//	}
//	return ErrorNoSuchProvider
//}
//
//func UpdateMediaUrl(media *model.Media) error {
//	if v, ok := Providers[media.Meta.(model.Meta).Name]; ok {
//		return v.UpdateMediaUrl(media)
//	}
//	return ErrorNoSuchProvider
//}
//
//func UpdateMediaLyric(media *model.Media) error {
//	if v, ok := Providers[media.Meta.(model.Meta).Name]; ok {
//		return v.UpdateMediaLyric(media)
//	}
//	return ErrorNoSuchProvider
//}
