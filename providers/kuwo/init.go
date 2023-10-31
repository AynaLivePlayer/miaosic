package kuwo

import "github.com/AynaLivePlayer/miaosic"

func init() {
	miaosic.RegisterProvider(NewKuwo())
}
