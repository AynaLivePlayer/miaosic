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
