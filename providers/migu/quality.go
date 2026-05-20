package migu

import (
	"strconv"
	"strings"

	"github.com/AynaLivePlayer/miaosic"
)

const (
	QualityLow      miaosic.Quality = "LQ"
	QualityStandard miaosic.Quality = "PQ"
	QualityHigh     miaosic.Quality = "HQ"
	QualityLossless miaosic.Quality = "SQ"
)

func (m *Migu) Qualities() []miaosic.Quality {
	return []miaosic.Quality{
		QualityLow,
		QualityStandard,
		QualityHigh,
		QualityLossless,
	}
}

var qualityMap = map[miaosic.Quality]miaosic.Quality{
	miaosic.QualityAny:      QualityHigh,
	miaosic.QualityStandard: QualityStandard,
	miaosic.Quality128k:     QualityStandard,
	miaosic.Quality192k:     QualityHigh,
	miaosic.Quality256k:     QualityHigh,
	miaosic.Quality320k:     QualityHigh,
	miaosic.QualityHQ:       QualityHigh,
	miaosic.QualitySQ:       QualityLossless,
	QualityLow:              QualityLow,
	QualityStandard:         QualityStandard,
	QualityHigh:             QualityHigh,
	QualityLossless:         QualityLossless,
}

func (m *Migu) MapQuality(quality miaosic.Quality) miaosic.Quality {
	if mapped, ok := qualityMap[quality]; ok {
		return mapped
	}
	upper := miaosic.Quality(strings.ToUpper(string(quality)))
	if mapped, ok := qualityMap[upper]; ok {
		return mapped
	}
	return QualityHigh
}

func pickMiguFormat(formats []miguRateFormat, quality miaosic.Quality, allowPaid bool) (miguRateFormat, bool) {
	if len(formats) == 0 {
		return miguRateFormat{}, false
	}

	type candidate struct {
		format miguRateFormat
		size   int64
		rank   int
	}
	candidates := make([]candidate, 0, len(formats))
	for _, format := range formats {
		if !allowPaid && isMiguPaidFormat(format) {
			continue
		}
		size, _ := strconv.ParseInt(firstNonZeroString(format.AndroidSize, format.ASize, format.Size, format.ISize), 10, 64)
		candidates = append(candidates, candidate{
			format: format,
			size:   size,
			rank:   miguFormatRank(format.FormatType),
		})
	}
	if len(candidates) == 0 {
		return miguRateFormat{}, false
	}

	requested := strings.ToUpper(string(quality))
	if requested != "" {
		var best *candidate
		for i := range candidates {
			if strings.ToUpper(strings.TrimSpace(candidates[i].format.FormatType)) != requested {
				continue
			}
			if best == nil || candidates[i].size > best.size {
				best = &candidates[i]
			}
		}
		if best != nil {
			return best.format, true
		}
	}

	best := candidates[0]
	for _, current := range candidates[1:] {
		if current.rank > best.rank || (current.rank == best.rank && current.size > best.size) {
			best = current
		}
	}
	return best.format, true
}

func isMiguPaidFormat(format miguRateFormat) bool {
	price, _ := strconv.Atoi(strings.TrimSpace(format.Price))
	if price >= 200 {
		return true
	}
	tags := format.ShowTag
	if len(tags) == 0 {
		tags = format.ShowTags
	}
	for _, tag := range tags {
		if strings.EqualFold(strings.TrimSpace(tag), "vip") {
			return true
		}
	}
	return false
}

func miguFormatRank(formatType string) int {
	switch strings.ToUpper(strings.TrimSpace(formatType)) {
	case "SQ":
		return 4
	case "ZQ":
		return 3
	case "HQ":
		return 2
	case "PQ":
		return 1
	case "LQ":
		return 0
	default:
		return -1
	}
}
