package kuwo

import "github.com/AynaLivePlayer/miaosic"

const (
	Quality128kMp3  miaosic.Quality = "128kmp3"
	Quality192kMp3  miaosic.Quality = "192kmp3"
	Quality256kMp3  miaosic.Quality = "256kmp3"
	Quality320kMp3  miaosic.Quality = "320kmp3"
	Quality2000Flac miaosic.Quality = "2000kflac"
)

func (k *Kuwo) Qualities() []miaosic.Quality {
	return []miaosic.Quality{
		Quality128kMp3,
		Quality192kMp3,
		Quality256kMp3,
		Quality320kMp3,
		Quality2000Flac,
	}
}

var qualityMap = map[miaosic.Quality]miaosic.Quality{
	miaosic.QualityAny:      Quality320kMp3,
	miaosic.QualityStandard: Quality128kMp3,
	miaosic.Quality128k:     Quality128kMp3,
	miaosic.Quality192k:     Quality192kMp3,
	miaosic.Quality256k:     Quality256kMp3,
	miaosic.Quality320k:     Quality320kMp3,
	miaosic.QualityHQ:       Quality320kMp3,
	miaosic.QualitySQ:       Quality2000Flac,
	Quality128kMp3:          Quality128kMp3,
	Quality192kMp3:          Quality192kMp3,
	Quality256kMp3:          Quality256kMp3,
	Quality320kMp3:          Quality320kMp3,
	Quality2000Flac:         Quality2000Flac,
}

func (k *Kuwo) MapQuality(quality miaosic.Quality) miaosic.Quality {
	if mapped, ok := qualityMap[quality]; ok {
		return mapped
	}
	return Quality320kMp3
}
