package cmds

import (
	"encoding/json"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/spf13/cobra"
	"strings"
)

var (
	searchPage     int
	searchPageSize int
)

func init() {
	CmdSearch.Flags().IntVarP(&searchPage, "page", "p", 1, "Page number")
	CmdSearch.Flags().IntVar(&searchPageSize, "page-size", 10, "Results per page")
}

var CmdSearch = &cobra.Command{
	Use:   "search <provider> <keyword>",
	Short: "Search media by keyword",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		keywords := args[1:]

		keyword := strings.Join(keywords, " ")

		provider, ok := miaosic.GetProvider(providerName)
		if !ok {
			fmt.Printf("Provider not found: %s\n", providerName)
			return
		}

		results, err := provider.Search(keyword, searchPage, searchPageSize)
		if err != nil {
			fmt.Printf("Error searching: %v\n", err)
			return
		}

		if len(results) == 0 {
			fmt.Println("No results found")
			return
		}

		result, _ := cmd.Flags().GetBool("json")
		if result {
			jsonBytes, err := json.Marshal(results)
			if err != nil {
				fmt.Printf("Error marshaling to JSON: %v\n", err)
				return
			}
			fmt.Println(string(jsonBytes))
			return
		}

		fmt.Printf("Page.%02d for \"%s\"\n", searchPage, keyword)
		for i, media := range results {
			fmt.Printf("%d. %s - %s - %s - %s\n",
				i+1,
				media.Title,
				media.Artist,
				media.Album,
				media.Meta.Identifier)
		}
	},
}
