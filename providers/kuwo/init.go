package kuwo

import "github.com/AynaLivePlayer/miaosic"

func Register() {
	miaosic.RegisterProvider(NewKuwo())
}
