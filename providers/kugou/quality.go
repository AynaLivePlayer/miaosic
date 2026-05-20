package kugou

import "github.com/AynaLivePlayer/miaosic"

const (
	// todo 添加魔法音质,  "dolby"
	Quality128k            miaosic.Quality = "128"         // 标准音质
	Quality320k            miaosic.Quality = "320"         // 高品音质
	QualityFlac            miaosic.Quality = "flac"        // 无损音质
	QualityHigh            miaosic.Quality = "high"        // Hi-Res音质
	QualityViperTape       miaosic.Quality = "viper_tape"  // 蝰蛇母带 少部分有 如果没有会返回320k
	QualityViperClear      miaosic.Quality = "viper_clear" // 蝰蛇超清
	QualityViperHiFi       miaosic.Quality = "viper_hifi"  // 蝰蛇hifi
	QualityViperAtmosphere miaosic.Quality = "viper_atmos" // 蝰蛇全景声
)

func (k *Kugou) Qualities() []miaosic.Quality {
	return []miaosic.Quality{
		Quality128k,
		Quality320k,
		QualityFlac,
		QualityHigh,
		QualityViperTape,
		QualityViperClear,
		QualityViperHiFi,
		QualityViperAtmosphere,
	}
}

func (k *KugouInstrumental) Qualities() []miaosic.Quality {
	return []miaosic.Quality{"magic_acappella"}
}

var qualityMap = map[miaosic.Quality]miaosic.Quality{
	miaosic.QualityAny:      Quality320k,
	miaosic.QualityStandard: Quality128k,
	miaosic.Quality128k:     Quality128k,
	miaosic.Quality192k:     Quality320k,
	miaosic.Quality256k:     Quality320k,
	miaosic.Quality320k:     Quality320k,
	miaosic.QualityHQ:       QualityHigh,
	miaosic.QualitySQ:       QualityFlac,
	Quality128k:             Quality128k,
	Quality320k:             Quality320k,
	QualityFlac:             QualityFlac,
	QualityHigh:             QualityHigh,
	QualityViperTape:        QualityViperTape,
	QualityViperClear:       QualityViperClear,
	QualityViperHiFi:        QualityViperHiFi,
	QualityViperAtmosphere:  QualityViperAtmosphere,
}

func (k *Kugou) MapQuality(quality miaosic.Quality) miaosic.Quality {
	if mapped, ok := qualityMap[quality]; ok {
		return mapped
	}
	if len(quality) > len("magic_") && string(quality[:len("magic_")]) == "magic_" {
		return quality
	}
	return Quality320k
}

func (k *KugouInstrumental) MapQuality(quality miaosic.Quality) miaosic.Quality {
	return "magic_acappella"
}
