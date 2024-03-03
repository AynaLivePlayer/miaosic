package netease

import "github.com/AynaLivePlayer/miaosic"

func init() {
	miaosic.RegisterProvider(NewNetease())
}
