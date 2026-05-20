package bilibili

import "github.com/AynaLivePlayer/miaosic"

func (b *Bilibili) Qualities() []miaosic.Quality {
	return []miaosic.Quality{miaosic.QualityAny}
}

var qualityMap = map[miaosic.Quality]miaosic.Quality{
	miaosic.QualityAny:      "2",
	miaosic.QualityStandard: "2",
	miaosic.Quality128k:     "2",
	miaosic.Quality192k:     "2",
	miaosic.Quality256k:     "2",
	miaosic.Quality320k:     "2",
	miaosic.QualityHQ:       "2",
	miaosic.QualitySQ:       "2",
}

func (b *Bilibili) MapQuality(quality miaosic.Quality) miaosic.Quality {
	if mapped, ok := qualityMap[quality]; ok {
		return mapped
	}
	return "2"
}
