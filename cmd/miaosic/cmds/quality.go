package cmds

import (
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

		fmt.Printf("Supported qualities for %s:\n", providerName)
		for i, quality := range qualities {
			fmt.Printf("%d. %s\n", i+1, quality)
		}
	},
}
