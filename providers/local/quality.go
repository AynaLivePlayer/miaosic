package local

import "github.com/AynaLivePlayer/miaosic"

func (l *Local) Qualities() []miaosic.Quality {
	return []miaosic.Quality{miaosic.QualityAny}
}

var qualityMap = map[miaosic.Quality]miaosic.Quality{
	miaosic.QualityAny:      miaosic.QualityStandard,
	miaosic.QualityStandard: miaosic.QualityStandard,
	miaosic.Quality128k:     miaosic.Quality128k,
	miaosic.Quality192k:     miaosic.Quality192k,
	miaosic.Quality256k:     miaosic.Quality256k,
	miaosic.Quality320k:     miaosic.Quality320k,
	miaosic.QualityHQ:       miaosic.QualityHQ,
	miaosic.QualitySQ:       miaosic.QualitySQ,
}

func (l *Local) MapQuality(quality miaosic.Quality) miaosic.Quality {
	if mapped, ok := qualityMap[quality]; ok {
		return mapped
	}
	return miaosic.QualityStandard
}
