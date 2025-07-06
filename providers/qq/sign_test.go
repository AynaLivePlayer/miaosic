package qq

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQqSignStr(t *testing.T) {
	data := "{\"module\":\"music.search.SearchCgiService\",\"method\":\"DoSearchForQQMusicMobile\",\"param\":{\"searchid\":\"xxx\",\"query\":\"asdfsadf\",\"search_type\":0,\"num_per_page\":10,\"page_num\":1,\"highlight\":1,\"grp\":1}}"
	require.Equal(t, "zzb226c4cd6u6x73owgk9ltzzy8yktygb3187a9d", qqSignStr(data))
}
