package bilibili

import (
	"miaosic"
)

func init() {
	miaosic.RegisterProvider(NewBilibili())
}
