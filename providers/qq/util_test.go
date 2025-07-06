package qq

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHash33(t *testing.T) {
	require.Equal(t, 1318936952, hash33("123fdas32asdf", 321))
}

func TestName(t *testing.T) {
	if time.UnixMilli(1756987630 * 1000).Before(time.Now().Add(24 * time.Hour)) {
		fmt.Println(123)
	}
}
