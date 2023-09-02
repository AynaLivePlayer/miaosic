package kuwo

import "miaosic"

func init() {
	miaosic.RegisterProvider(NewKuwo(miaosic.Requester))
}
