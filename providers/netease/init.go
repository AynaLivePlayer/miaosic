package netease

import "github.com/AynaLivePlayer/miaosic"

func Register() {
	miaosic.RegisterProvider(NewNetease())
}
