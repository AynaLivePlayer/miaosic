package bilivideo

import "github.com/AynaLivePlayer/miaosic"

func init() {
	miaosic.RegisterProvider(NewBilibiliViedo())
}
