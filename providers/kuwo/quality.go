package kuwo

import "github.com/AynaLivePlayer/miaosic"

const (
	// convert_url2 format=flac: 可以拿到可播放 flac URL，实测返回 bitrate$2000&format$flac。
	QualityFlac miaosic.Quality = "flac"
	// convert_url2 format=mp3: 可以拿到可播放 mp3 URL，实测返回 bitrate$320&format$mp3。
	QualityMp3 miaosic.Quality = "mp3"
	// convert_url2 format=128kmp3: 可以拿到加密资源 URL；当前可播放流程映射到 format=mp3。
	Quality128kMp3 miaosic.Quality = "128kmp3"
	// convert_url2 format=192kmp3: 可以拿到加密资源 URL；当前可播放流程映射到 format=mp3。
	Quality192kMp3 miaosic.Quality = "192kmp3"
	// 不在本次探测的加密格式列表中；当前可播放流程映射到 format=mp3。
	Quality256kMp3 miaosic.Quality = "256kmp3"
	// convert_url2 format=320kmp3: 也可以拿到加密资源 URL。
	Quality320kMp3 miaosic.Quality = "320kmp3"
	// convert_url2 format=2000kflac: 可以拿到加密资源 URL；当前可播放流程映射到 format=flac。
	Quality2000Flac miaosic.Quality = "2000kflac"
	// convert_url2 format=4000kflac: 可以拿到加密资源 URL；当前可播放流程映射到 format=flac。
	Quality4000Flac miaosic.Quality = "4000kflac"
)

func (k *Kuwo) Qualities() []miaosic.Quality {
	return []miaosic.Quality{
		QualityMp3,
		QualityFlac,
	}
}

var qualityMap = map[miaosic.Quality]miaosic.Quality{
	miaosic.QualityAny:      QualityMp3,
	miaosic.QualityStandard: QualityMp3,
	miaosic.Quality128k:     QualityMp3,
	miaosic.Quality192k:     QualityMp3,
	miaosic.Quality256k:     QualityMp3,
	miaosic.Quality320k:     QualityMp3,
	miaosic.QualityHQ:       QualityMp3,
	miaosic.QualitySQ:       QualityFlac,
	QualityFlac:             QualityFlac,
	QualityMp3:              QualityMp3,
	Quality128kMp3:          QualityMp3,
	Quality192kMp3:          QualityMp3,
	Quality256kMp3:          QualityMp3,
	Quality320kMp3:          QualityMp3,
	Quality2000Flac:         QualityFlac,
	Quality4000Flac:         QualityFlac,
}

func (k *Kuwo) MapQuality(quality miaosic.Quality) miaosic.Quality {
	if mapped, ok := qualityMap[quality]; ok {
		return mapped
	}
	return QualityMp3
}
