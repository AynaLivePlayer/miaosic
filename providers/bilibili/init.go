package bilibili

import (
	"github.com/AynaLivePlayer/miaosic"
)

func Register() {
	miaosic.RegisterProvider(NewBilibili())
}
