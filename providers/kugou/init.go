package kugou

import "github.com/AynaLivePlayer/miaosic"

var api *Kugou

func init() {
	api = NewKugou(false)
	miaosic.RegisterProvider(api)
}

func UseInstrumental() {
	miaosic.RegisterProvider(&KugouInstrumental{api})
}
