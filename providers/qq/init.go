package qq

import (
	"github.com/AynaLivePlayer/miaosic"
	"math/rand"
	"time"
)

func init() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	// user should manually register provider since there are two channel
	//miaosic.RegisterProvider(NewQQMusicProvider("qq"))
}

func UseQQLogin() {
	miaosic.RegisterProvider(NewQQMusicProvider("qq"))
}

func UseWechatLogin() {
	miaosic.RegisterProvider(NewQQMusicProvider("wechat"))
}
