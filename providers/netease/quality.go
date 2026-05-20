package netease

import "github.com/AynaLivePlayer/miaosic"

// todo  jyeffect => 高清环绕声, Audio Vivid, 高清臻音(spatial audio)

// 如果不存在会获取最高的
const (
	// 音质
	QualityStandard miaosic.Quality = "standard" // 标准 128kbps vip
	QualityHigher   miaosic.Quality = "higher"   // 较高 vip
	QualityExHigh   miaosic.Quality = "exhigh"   // 极高(HQ) 最高320kbps vip
	QualityLossless miaosic.Quality = "lossless" // 无损(SQ) 最高 48kHz/16bit vip
	QualityHiRes    miaosic.Quality = "hires"    // 高解析度无损(Hi-Res) 最高192kHz/24bit vip
	QualityJyMaster miaosic.Quality = "jymaster" // 超清母带 192kHz/24bit svip
	// 空间音感
	QualitySky miaosic.Quality = "sky" // 沉浸环绕声 Surround Audio svip
)

func (n *Netease) Qualities() []miaosic.Quality {
	return []miaosic.Quality{
		QualityStandard,
		QualityHigher,
		QualityExHigh,
		QualityLossless,
		QualityHiRes,
		QualityJyMaster,
	}
}

var qualityMap = map[miaosic.Quality]miaosic.Quality{
	miaosic.QualityAny:      QualityStandard,
	miaosic.QualityStandard: QualityStandard,
	miaosic.Quality128k:     QualityStandard,
	miaosic.Quality192k:     QualityHigher,
	miaosic.Quality256k:     QualityHigher,
	miaosic.Quality320k:     QualityExHigh,
	miaosic.QualityHQ:       QualityExHigh,
	miaosic.QualitySQ:       QualityLossless,
	QualityHigher:           QualityHigher,
	QualityExHigh:           QualityExHigh,
	QualityLossless:         QualityLossless,
	QualityHiRes:            QualityHiRes,
	QualityJyMaster:         QualityJyMaster,
	QualitySky:              QualitySky,
}

func (n *Netease) MapQuality(quality miaosic.Quality) miaosic.Quality {
	if mapped, ok := qualityMap[quality]; ok {
		return mapped
	}
	return QualityStandard
}
