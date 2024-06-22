package kugou

import "github.com/AynaLivePlayer/miaosic"

func init() {
	miaosic.RegisterProvider(NewKugou())
}
