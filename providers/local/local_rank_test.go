package local

import (
	"testing"

	"github.com/AynaLivePlayer/miaosic"
	"github.com/stretchr/testify/require"
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
	result := rankMedia("王菲", &testData)
	require.NotEmpty(t, result)
	require.Equal(t, "王菲", result[0].Artist)
}

func TestLocal_SearchTest2(t *testing.T) {
	result := rankMedia("怪物 reol", &testData)
	require.NotEmpty(t, result)
	require.Equal(t, "怪物", result[0].Title)
	require.Equal(t, "reol", result[0].Artist)
}

func TestLocal_RankEmptyKeywordReturnsAll(t *testing.T) {
	result := rankMedia("", &testData)
	require.Len(t, result, len(testData))
	require.Equal(t, testData[0], result[0])
}
