package utils

import (
	"github.com/AynaLivePlayer/miaosic"
	"github.com/abadojack/whatlanggo"
)

func DetectSetLyricLang(lyric *miaosic.Lyrics) string {
	lyrics := ""
	for _, lrcLine := range lyric.Content {
		lyrics += " " + lrcLine.Lyric
	}
	lang := whatlanggo.DetectLang(lyrics)
	lyric.Lang = lang.Iso6393()
	return lyrics
}

func ParseLyricWithLangDetection(lyrics string) miaosic.Lyrics {
	lyric := miaosic.ParseLyrics("default", lyrics)
	DetectSetLyricLang(&lyric)
	return lyric
}
