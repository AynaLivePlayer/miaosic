package miaosic

import (
	"github.com/spf13/cast"
	"regexp"
	"sort"
	"strings"
)

var timeTagRegex = regexp.MustCompile("\\[[0-9]+:[0-9]+(\\.[0-9]+)?\\]")

type LyricLine struct {
	Time  float64 // in seconds
	Lyric string
}

type Lyrics struct {
	Lang    string
	Content []LyricLine
}

func ParseLyrics(lang string, lyrics string) Lyrics {
	tmp := make(map[float64]LyricLine)
	times := make([]float64, 0)
	for _, line := range strings.Split(lyrics, "\n") {
		lrc := timeTagRegex.ReplaceAllString(line, "")
		if len(lrc) > 0 && lrc[len(lrc)-1] == '\r' {
			lrc = lrc[:len(lrc)-1]
		}
		for _, time := range timeTagRegex.FindAllString(line, -1) {
			ts := strings.Split(time[1:len(time)-1], ":")
			t := cast.ToFloat64(ts[0])*60 + cast.ToFloat64(ts[1])
			times = append(times, t)
			tmp[t] = LyricLine{
				Time:  t,
				Lyric: lrc,
			}
		}
	}
	sort.Float64s(times)
	lrcs := make([]LyricLine, len(times))
	for index, time := range times {
		lrcs[index] = tmp[time]
	}
	if len(lrcs) == 0 {
		lrcs = append(lrcs, LyricLine{Time: 0, Lyric: ""})
	}
	lrcs = append(lrcs, LyricLine{
		Time: lrcs[len(lrcs)-1].Time + 5,
	})
	lrcs = append(lrcs, LyricLine{
		Time:  99999999999,
		Lyric: "",
	})
	return Lyrics{Lang: lang, Content: lrcs}
}

func (l Lyrics) FindIndex(time float64) int {
	start := 0
	end := len(l.Content) - 1
	mid := (start + end) / 2
	for start < end {
		if l.Content[mid].Time <= time && time < l.Content[mid+1].Time {
			return mid
		}
		if l.Content[mid].Time > time {
			end = mid
		} else {
			start = mid
		}
		mid = (start + end) / 2
	}
	return -1
}

func (l Lyrics) Find(time float64) LyricLine {
	idx := l.FindIndex(time)
	if idx == -1 {
		return LyricLine{Time: 0, Lyric: ""}
	}
	return l.Content[idx]
}
