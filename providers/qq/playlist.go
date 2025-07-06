package qq

import (
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"regexp"
	"strings"
)

var playlistRegexp = regexp.MustCompile(`ryqq/playlist/(\d+)`)
var playlistRegexp0 = regexp.MustCompile("^[0-9]+$")

func (p *QQMusicProvider) MatchPlaylist(uri string) (miaosic.MetaData, bool) {
	if id := playlistRegexp0.FindString(uri); id != "" {
		return miaosic.MetaData{
			Provider:   p.GetName(),
			Identifier: id,
		}, true
	}
	if id := playlistRegexp.FindString(uri); id != "" {
		return miaosic.MetaData{
			Provider:   p.GetName(),
			Identifier: id[14:],
		}, true
	}
	return miaosic.MetaData{}, false
}

func (p *QQMusicProvider) getPlaylist(meta miaosic.MetaData, page, pagesize int) (gjson.Result, error) {
	params := map[string]interface{}{
		"disstid":      cast.ToInt(meta.Identifier),
		"dirid":        0,
		"tag":          0,
		"song_begin":   pagesize * (page - 1),
		"song_num":     pagesize,
		"userinfo":     0,
		"orderlist":    1,
		"onlysonglist": 0,
	}
	return p.makeApiRequest("music.srfDissInfo.DissInfo", "CgiGetDiss", params)
}

func (p *QQMusicProvider) GetPlaylist(meta miaosic.MetaData) (*miaosic.Playlist, error) {
	// todo use RequestGroup.
	playlist := &miaosic.Playlist{
		Meta:   meta,
		Title:  "QQPlaylist " + meta.Identifier,
		Medias: make([]miaosic.MediaInfo, 0),
	}
	for page := 1; page < 20; page++ {
		data, err := p.getPlaylist(meta, page, 100)
		if err != nil {
			return nil, err
		}
		totalSongNum := int(data.Get("data.total_song_num").Int())
		data.Get("data.songlist").ForEach(func(k, info gjson.Result) bool {
			albumMid := info.Get("album.mid").String()
			coverURL := ""
			if albumMid != "" {
				coverURL = fmt.Sprintf("https://y.qq.com/music/photo_new/T002R500x500M000%s.jpg", albumMid)
			}
			var artistNames []string
			info.Get("singer").ForEach(func(_, singer gjson.Result) bool {
				name := singer.Get("name").String()
				if name != "" {
					artistNames = append(artistNames, name)
				}
				return true
			})
			playlist.Medias = append(playlist.Medias, miaosic.MediaInfo{
				Title:  info.Get("title").String(),
				Artist: strings.Join(artistNames, ","),
				Album:  info.Get("album.title").String(),
				Cover:  miaosic.Picture{Url: coverURL},
				Meta:   miaosic.MetaData{Provider: p.GetName(), Identifier: info.Get("mid").String()},
			})
			return true
		})
		if page == 1 {
			playlist.Title = data.Get("data.dirinfo.title").String()
		}
		if len(playlist.Medias) >= totalSongNum {
			break
		}
	}
	return playlist, nil
}
