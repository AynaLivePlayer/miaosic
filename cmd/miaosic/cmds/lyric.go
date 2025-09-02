package cmds

import (
	"encoding/json"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

var (
	lyricOutput string
	saveLyric   bool
)

func init() {
	CmdLyric.Flags().StringVarP(&lyricOutput, "output", "o", "", "Output lyrics to file")
	CmdLyric.Flags().BoolVar(&saveLyric, "save", false, "Save lyrics to file with auto-generated name")
	CmdLyric.Flags().BoolP("json", "j", false, "output in json")
}

var CmdLyric = &cobra.Command{
	Use:   "lyric <provider> <uri>",
	Short: "Get media lyrics",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		uri := args[1]

		provider, ok := miaosic.GetProvider(providerName)
		if !ok {
			fmt.Printf("Provider not found: %s\n", providerName)
			return
		}

		meta, ok := provider.MatchMedia(uri)
		if !ok {
			fmt.Printf("URI not matched by provider: %s\n", uri)
			return
		}

		lyrics, err := provider.GetMediaLyric(meta)
		if err != nil {
			fmt.Printf("Error getting media lyrics: %v\n", err)
			return
		}

		if len(lyrics) == 0 {
			fmt.Println("No lyrics found")
			return
		}

		result, _ := cmd.Flags().GetBool("json")
		if result {
			jsonBytes, err := json.Marshal(lyrics)
			if err != nil {
				fmt.Printf("Error marshaling to JSON: %v\n", err)
				return
			}
			fmt.Println(string(jsonBytes))
			return
		}

		var mediaInfo miaosic.MediaInfo
		if saveLyric && lyricOutput == "" {
			info, err := provider.GetMediaInfo(meta)
			if err != nil {
				fmt.Printf("Failed to get media info for filename: %v\n", err)
				return
			}
			mediaInfo = info
		}

		outputToFile := lyricOutput != "" || saveLyric

		if outputToFile {
			// 确定基础文件名
			baseFilename := lyricOutput
			if baseFilename == "" {
				// 生成基于媒体信息的文件名
				title := sanitizeFilename(mediaInfo.Title)
				artist := sanitizeFilename(mediaInfo.Artist)
				if title == "" {
					title = "unknown_title"
				}
				if artist == "" {
					artist = "unknown_artist"
				}
				baseFilename = fmt.Sprintf("%s_%s.lrc", title, artist)
			}

			if baseFilename == "" {
				baseFilename = "lyrics.lrc"
			}

			// 处理多语言歌词
			for _, lyric := range lyrics {
				lang := lyric.Lang
				if lang == "" {
					lang = "unknown"
				}

				var filename string
				if len(lyrics) == 1 {
					filename = baseFilename
				} else {
					ext := filepath.Ext(baseFilename)
					base := strings.TrimSuffix(baseFilename, ext)
					filename = fmt.Sprintf("%s_%s%s", base, lang, ext)
				}

				// 写入文件
				if err := os.WriteFile(filename, []byte(lyric.String()), 0644); err != nil {
					fmt.Printf("Failed to write lyrics to %s: %v\n", filename, err)
				} else {
					fmt.Printf("Lyrics saved to: %s\n", filename)
				}
			}
		} else {
			// 输出到控制台
			if len(lyrics) == 0 {
				fmt.Println("No lyrics found")
				return
			}

			for _, lyric := range lyrics {
				lang := lyric.Lang
				if lang == "" {
					lang = "unknown"
				}

				fmt.Printf("Language: %s\n", lang)
				fmt.Println("-----")
				fmt.Println(lyric.String())
				fmt.Println("-----")
			}
		}
	},
}
