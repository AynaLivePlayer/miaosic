package local

import (
	"sort"
	"strings"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/sahilm/fuzzy"
)

type mediaRanking struct {
	media miaosic.MediaInfo
	score int
	order int
}

type localSearchDoc struct {
	info   miaosic.MediaInfo
	search string
}

func localSearchText(info miaosic.MediaInfo) string {
	parts := []string{
		info.Title,
		info.Artist,
		info.Album,
		info.Meta.Identifier,
	}
	parts = append(parts, info.Artists...)
	return strings.ToLower(strings.Join(parts, " "))
}

func rankMedia(keyword string, medias *[]miaosic.MediaInfo) []miaosic.MediaInfo {
	docs := make([]localSearchDoc, 0, len(*medias))
	for _, media := range *medias {
		docs = append(docs, localSearchDoc{
			info:   media,
			search: localSearchText(media),
		})
	}
	return rankLocalSearchDocs(keyword, docs)
}

func rankLocalSearchDocs(keyword string, docs []localSearchDoc) []miaosic.MediaInfo {
	patterns := strings.Fields(keyword)
	if len(patterns) == 0 {
		result := make([]miaosic.MediaInfo, 0, len(docs))
		for _, doc := range docs {
			result = append(result, doc.info)
		}
		return result
	}

	data := make([]mediaRanking, 0, len(docs))
	dataStr := make([]string, 0, len(docs))
	for i, doc := range docs {
		data = append(data, mediaRanking{
			media: doc.info,
			order: i,
		})
		dataStr = append(dataStr, doc.search)
	}
	for _, pattern := range patterns {
		pattern = strings.ToLower(pattern)
		for i, search := range dataStr {
			if strings.Contains(search, pattern) {
				data[i].score += len(pattern) * 100
			}
		}
		for _, match := range fuzzy.Find(pattern, dataStr) {
			data[match.Index].score += match.Score
		}
	}

	sort.Slice(data, func(i, j int) bool {
		if data[i].score == data[j].score {
			return data[i].order < data[j].order
		}
		return data[i].score > data[j].score
	})

	result := make([]miaosic.MediaInfo, 0, len(data))
	for _, d := range data {
		if d.score > 0 {
			result = append(result, d.media)
		}
	}
	return result
}
