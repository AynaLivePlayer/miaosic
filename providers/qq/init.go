package qq

import (
	"github.com/AynaLivePlayer/miaosic"
	"math/rand"
	"time"
)

func init() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	miaosic.RegisterProvider(NewQQMusicProvider())
}
