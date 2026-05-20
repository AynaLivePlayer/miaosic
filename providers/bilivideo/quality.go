package bilivideo

import "github.com/AynaLivePlayer/miaosic"

func (b *BilibiliVideo) Qualities() []miaosic.Quality {
	return []miaosic.Quality{miaosic.QualityAny}
}

var qualityMap = map[miaosic.Quality]miaosic.Quality{
	miaosic.QualityAny:      "32",
	miaosic.QualityStandard: "16",
	miaosic.Quality128k:     "16",
	miaosic.Quality192k:     "32",
	miaosic.Quality256k:     "32",
	miaosic.Quality320k:     "32",
	miaosic.QualityHQ:       "32",
	miaosic.QualitySQ:       "32",
}

func (b *BilibiliVideo) MapQuality(quality miaosic.Quality) miaosic.Quality {
	if mapped, ok := qualityMap[quality]; ok {
		return mapped
	}
	return "32"
}
