package cmds

import (
	"encoding/json"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/spf13/cobra"
)

func init() {
	CmdInfo.Flags().BoolP("json", "j", false, "output in json")
}

var CmdInfo = &cobra.Command{
	Use:   "info <provider> <uri>",
	Short: "Get media info",
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

		info, err := provider.GetMediaInfo(meta)
		if err != nil {
			fmt.Printf("Error getting media info: %v\n", err)
			return
		}
		result, _ := cmd.Flags().GetBool("json")
		if result {
			jsonBytes, err := json.Marshal(info)
			if err != nil {
				fmt.Printf("Error marshaling to JSON: %v\n", err)
				return
			}
			fmt.Println(string(jsonBytes))
			return
		}

		fmt.Println("Title:", info.Title)
		fmt.Println("Artist:", info.Artist)
		fmt.Println("Album:", info.Album)
		fmt.Println("Cover", info.Cover.Url)
		fmt.Println("Provider:", info.Meta.Provider)
		fmt.Println("Identifier:", info.Meta.Identifier)
	},
}
