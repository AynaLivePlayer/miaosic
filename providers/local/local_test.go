package local

import (
	"fmt"
	"github.com/sahilm/fuzzy"
	"miaosic"
	"sort"
	"strings"
	"testing"
)

var testData = []miaosic.MediaInfo{
	{Title: "Shape of You", Artist: "Ed Sheeran"},
	{Title: "Lose Yourself", Artist: "Eminem"},
	{Title: "Believer", Artist: "Imagine Dragons"},
	{Title: "Counting Stars", Artist: "OneRepublic"},
	{Title: "Rolling in the Deep", Artist: "Adele"},
	{Title: "Uptown Funk", Artist: "Mark Ronson ft. Bruno Mars"},
	{Title: "Imagine", Artist: "John Lennon"},
	{Title: "I Will Always Love You", Artist: "Whitney Houston"},
	{Title: "Smells Like Teen Spirit", Artist: "Nirvana"},
	{Title: "Billie Jean", Artist: "Michael Jackson"},

	// Chinese songs
	{Title: "平凡之路", Artist: "朴树"},
	{Title: "染", Artist: "reol"},
	{Title: "怪物", Artist: "reol"},
	{Title: "怪物", Artist: "王菲"},
	{Title: "怪物", Artist: "怪物"},
	{Title: "小幸运", Artist: "田馥甄"},
	{Title: "遥远的她", Artist: "张学友"},
	{Title: "匆匆那年", Artist: "王菲"},
	{Title: "岁月神偷", Artist: "金玟岐"},
	{Title: "突然好想你", Artist: "五月天"},
	{Title: "蓝莲花", Artist: "许巍"},
	{Title: "红豆", Artist: "王菲"},
	{Title: "夜空中最亮的星", Artist: "逃跑计划"},
	{Title: "爱情转移", Artist: "陈奕迅"},
}

func TestLocal_SearchTest1(t *testing.T) {
	testPattern := "王菲"
	patterns := strings.Split(testPattern, " ")
	data := make([]*mediaRanking, 0)

	for _, media := range testData {
		m := media
		data = append(data, &mediaRanking{
			media: &m,
			score: 0,
		})
	}
	dataStr := make([]string, 0)
	for _, d := range data {
		dataStr = append(dataStr, strings.ToLower(d.media.Title+" "+d.media.Artist))
	}

	for _, pattern := range patterns {
		for _, match := range fuzzy.Find(pattern, dataStr) {
			data[match.Index].score += match.Score
		}
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].score > data[j].score
	})

	for _, d := range data {
		fmt.Println(d.score, d.media)
	}
}

func TestLocal_SearchTest2(t *testing.T) {
	for _, media := range rankMedia("怪物 reol", &testData) {
		fmt.Println(media)
	}
}
