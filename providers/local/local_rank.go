package local

import (
	"github.com/AynaLivePlayer/miaosic"
	"github.com/sahilm/fuzzy"
	"sort"
	"strings"
)

type mediaRanking struct {
	media *miaosic.MediaInfo
	score int
}

func rankMedia(keyword string, medias *[]miaosic.MediaInfo) []miaosic.MediaInfo {
	patterns := strings.Split(keyword, " ")
	data := make([]*mediaRanking, 0)

	for i, _ := range *medias {
		data = append(data, &mediaRanking{
			media: &(*medias)[i],
			score: 0,
		})
	}

	for _, pattern := range patterns {
		pattern = strings.ToLower(pattern)
		dataStr := make([]string, 0)
		for _, d := range data {
			dataStr = append(dataStr, strings.ToLower(d.media.Title))
		}
		for _, match := range fuzzy.Find(pattern, dataStr) {
			data[match.Index].score += match.Score
		}
		dataStr = make([]string, 0)
		for _, d := range data {
			dataStr = append(dataStr, strings.ToLower(d.media.Artist))
		}
		for _, match := range fuzzy.Find(pattern, dataStr) {
			data[match.Index].score += match.Score
		}
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].score > data[j].score
	})

	result := make([]miaosic.MediaInfo, 0)
	for _, d := range data {
		if d.score > 0 {
			result = append(result, *d.media)
		}
	}
	return result
}
