package kugou

import "github.com/AynaLivePlayer/miaosic"

var api = NewKugou(false)

func Register() {
	miaosic.RegisterProvider(api)
}

func UseInstrumental() {
	miaosic.RegisterProvider(&KugouInstrumental{api})
}
