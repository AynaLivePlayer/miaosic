package cmds

import (
	"encoding/json"
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/spf13/cobra"
	"os"
)

var CmdQuality = &cobra.Command{
	Use:   "quality <provider>",
	Short: "List supported qualities for a provider",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]

		provider, ok := miaosic.GetProvider(providerName)
		if !ok {
			fmt.Printf("Provider not found: %s\n", providerName)
			os.Exit(1)
		}

		qualities := provider.Qualities()
		if len(qualities) == 0 {
			fmt.Printf("No qualities found for provider: %s\n", providerName)
			return
		}

		result, _ := cmd.Flags().GetBool("json")
		if result {
			jsonBytes, err := json.Marshal(qualities)
			if err != nil {
				fmt.Printf("Error marshaling to JSON: %v\n", err)
				return
			}
			fmt.Println(string(jsonBytes))
			return
		}

		fmt.Printf("Supported qualities for %s:\n", providerName)
		for i, quality := range qualities {
			fmt.Printf("%d. %s\n", i+1, quality)
		}
	},
}
