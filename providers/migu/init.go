package migu

import "github.com/AynaLivePlayer/miaosic"

func Register() {
	miaosic.RegisterProvider(NewMigu())
}
