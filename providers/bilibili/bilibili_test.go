package bilibili

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"miaosic"
	"testing"
)

var api miaosic.MediaProvider = NewBilibili(miaosic.Requester)

func TestBilibili_Search(t *testing.T) {
	result, err := api.Search("染 reol")
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

func TestBilibili_GetMusic(t *testing.T) {
	media := miaosic.Media{
		Meta: miaosic.MediaMeta{
			Provider:   api.GetName(),
			Identifier: "1560601",
		},
	}
	require.NoError(t, api.UpdateMedia(&media))
	require.NoError(t, api.UpdateMediaUrl(&media))
	fmt.Println(media.Url)
}
