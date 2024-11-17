package kugou

import (
	"github.com/AynaLivePlayer/miaosic"
)

type KugouInstrumental struct {
	k *Kugou
}

func (k *KugouInstrumental) GetName() string {
	return "Kugou-Instr"
}

func (k *KugouInstrumental) Search(keyword string, page, size int) ([]miaosic.MediaInfo, error) {
	result, err := k.k.Search(keyword, page, size)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(result); i++ {
		result[i].Meta.Provider = k.GetName()
	}
	return result, nil
}

func (k *KugouInstrumental) MatchMedia(uri string) (miaosic.MetaData, bool) {
	m, ok := k.k.MatchMedia(uri)
	if !ok {
		return m, ok
	}
	m.Provider = k.GetName()
	return m, ok
}

func (k *KugouInstrumental) GetMediaInfo(meta miaosic.MetaData) (miaosic.MediaInfo, error) {
	m, err := k.k.GetMediaInfo(meta)
	if err != nil {
		return m, err
	}
	m.Meta.Provider = k.GetName()
	return m, nil
}

func (k *KugouInstrumental) GetMediaUrl(meta miaosic.MetaData, quality miaosic.Quality) ([]miaosic.MediaUrl, error) {
	return k.k.GetMediaUrl(meta, "magic_acappella")
}

func (k *KugouInstrumental) GetMediaLyric(meta miaosic.MetaData) ([]miaosic.Lyrics, error) {
	return k.k.GetMediaLyric(meta)
}

func (k *KugouInstrumental) MatchPlaylist(uri string) (miaosic.MetaData, bool) {
	m, ok := k.k.MatchPlaylist(uri)
	if !ok {
		return m, ok
	}
	m.Provider = k.GetName()
	return m, ok
}

func (k *KugouInstrumental) GetPlaylist(meta miaosic.MetaData) (*miaosic.Playlist, error) {
	p, err := k.k.GetPlaylist(meta)
	if err != nil {
		return p, err
	}
	p.Meta.Provider = k.GetName()
	for i := 0; i < len(p.Medias); i++ {
		p.Medias[i].Meta.Provider = k.GetName()
	}
	return p, nil
}
