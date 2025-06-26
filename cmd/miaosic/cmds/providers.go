package cmds

import (
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/spf13/cobra"
)

var CmdProviders = &cobra.Command{
	Use:   "providers",
	Short: "List all registered providers and login status",
	Run: func(cmd *cobra.Command, args []string) {
		providers := miaosic.ListAvailableProviders()

		if len(providers) == 0 {
			fmt.Println("No providers registered")
			return
		}

		for _, providerName := range providers {
			fmt.Printf(" - %s: ", providerName)
			provider, _ := miaosic.GetProvider(providerName)
			// 检查登录状态
			if loginable, ok := provider.(miaosic.Loginable); ok {
				status := "Not logged in"
				if loginable.IsLogin() {
					status = "Logged in"
				}
				fmt.Printf("%s\n", status)
			} else {
				fmt.Println("Not supported")
			}
		}
	},
}
