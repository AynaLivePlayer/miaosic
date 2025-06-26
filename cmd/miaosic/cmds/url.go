package cmds

import (
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/spf13/cobra"
)

func init() {
	CmdUrl.Flags().String("quality", "", "Quality preference (128k, 192k, 256k, 320k, hq, sq)")
}

var CmdUrl = &cobra.Command{
	Use:   "url <provider> <uri>",
	Short: "Get media URLs",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		uri := args[1]
		quality, _ := cmd.Flags().GetString("quality")

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

		urls, err := provider.GetMediaUrl(meta, miaosic.Quality(quality))
		if err != nil {
			fmt.Printf("Error getting media URLs: %v\n", err)
			return
		}

		for i, url := range urls {
			fmt.Printf("URL %d:\n", i+1)
			fmt.Printf("  Quality: %s\n", url.Quality)
			fmt.Printf("  URL: %s\n", url.Url)
			if len(url.Header) > 0 {
				fmt.Println("  Headers:")
				for k, v := range url.Header {
					fmt.Printf("    %s: %s\n", k, v)
				}
			}
			fmt.Println()
		}
	},
}
