package qq

import (
	"math/rand"
	"time"

	"github.com/AynaLivePlayer/miaosic"
)

func init() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	// user should manually register provider since there are two channel
	//miaosic.RegisterProvider(NewQQMusicProvider("qq"))
}

func Register() {
	UseQQLogin()
}

func UseQQLogin() {
	miaosic.RegisterProvider(NewQQMusicProvider("qq"))
}

func UseWechatLogin() {
	miaosic.RegisterProvider(NewQQMusicProvider("wechat"))
}
