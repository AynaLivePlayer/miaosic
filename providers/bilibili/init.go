package bilibili

import (
	"github.com/AynaLivePlayer/miaosic"
)

func init() {
	miaosic.RegisterProvider(NewBilibili())
}
