package main

import (
	"fmt"
	"github.com/AynaLivePlayer/miaosic/cmd/miaosic/cmds"
	"github.com/AynaLivePlayer/miaosic/cmd/miaosic/internal"
	_ "github.com/AynaLivePlayer/miaosic/providers/bilivideo"
	"github.com/AynaLivePlayer/miaosic/providers/kugou"
	_ "github.com/AynaLivePlayer/miaosic/providers/kugou"
	_ "github.com/AynaLivePlayer/miaosic/providers/kuwo"
	_ "github.com/AynaLivePlayer/miaosic/providers/local"
	_ "github.com/AynaLivePlayer/miaosic/providers/netease"
	"github.com/AynaLivePlayer/miaosic/providers/qq"
	"github.com/spf13/cobra"
)

func init() {
	kugou.UseInstrumental()
	qq.UseQQLogin()
}

var rootCmd = &cobra.Command{
	Use:   "miaosic",
	Short: "cmdline tool for miaosic.",
	Long:  `cmdline tool for miaosic: a music searching tools`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := internal.RestoreSessions(sessionFile); err != nil {
			fmt.Printf("Error restoring sessions from file: %v\n", err)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if err := internal.SaveSessions(sessionFile); err != nil {
			fmt.Printf("Error saving sessions: %v\n", err)
		}
	},
}

var sessionFile string

func init() {
	rootCmd.PersistentFlags().StringVarP(&sessionFile, "session-file", "s", "", "Session file path")
	rootCmd.AddCommand(cmds.CmdProviders)
	rootCmd.AddCommand(cmds.CmdSearch)
	rootCmd.AddCommand(cmds.CmdQrlogin)
	rootCmd.AddCommand(cmds.CmdInfo)
	rootCmd.AddCommand(cmds.CmdUrl)
	rootCmd.AddCommand(cmds.CmdLyric)
	rootCmd.AddCommand(cmds.CmdQuality)
	rootCmd.AddCommand(cmds.CmdDownload)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
