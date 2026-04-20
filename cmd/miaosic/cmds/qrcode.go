package cmds

import (
	"fmt"
	"github.com/AynaLivePlayer/miaosic"
	"github.com/AynaLivePlayer/miaosic/cmd/miaosic/internal"
	"github.com/spf13/cobra"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/file"
	"os"
)

var CmdQrlogin = &cobra.Command{
	Use:   "qrlogin",
	Short: "QR code login operations",
	Long:  "Manage provider login sessions using QR code flow.",
}

var getqrcodeCmd = &cobra.Command{
	Use:     "getqrcode <provider>",
	Short:   "Get QR code for login",
	Long:    "Generate and print a login QR code, then return key/url for verification.",
	Example: "  miaosic qrlogin getqrcode netease\n  miaosic qrlogin getqrcode qq",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]

		provider, ok := miaosic.GetProvider(providerName)
		if !ok {
			fmt.Printf("Provider not found: %s\n", providerName)
			return
		}

		loginable, ok := provider.(miaosic.Loginable)
		if !ok {
			fmt.Printf("Provider does not support login: %s\n", providerName)
			return
		}

		qrSession, err := loginable.QrLogin()
		if err != nil {
			fmt.Printf("Error getting QR code: %v\n", err)
			return
		}

		qrc, err := qrcode.New(qrSession.Url)
		if err != nil {
			fmt.Printf("Error creating QR code: %v\n", err)
			return
		}

		w := file.New(os.Stdout)
		fmt.Println("Scan this QR code to login:")
		if err := qrc.Save(w); err != nil {
			fmt.Printf("Error printing QR code: %v\n", err)
		}

		fmt.Println("Key:", qrSession.Key)
		fmt.Println("URL:", qrSession.Url)
	},
}

var verifyCmd = &cobra.Command{
	Use:     "verify <provider> <key>",
	Short:   "Verify QR login",
	Long:    "Verify a scanned QR login key and persist the provider session.",
	Example: "  miaosic qrlogin verify netease <key>\n  miaosic --session-file ~/.miaosic_session.json qrlogin verify qq <key>",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		key := args[1]

		provider, ok := miaosic.GetProvider(providerName)
		if !ok {
			fmt.Printf("Provider not found: %s\n", providerName)
			return
		}

		loginable, ok := provider.(miaosic.Loginable)
		if !ok {
			fmt.Printf("Provider does not support login: %s\n", providerName)
			return
		}

		qrSession := &miaosic.QrLoginSession{Key: key}
		result, err := loginable.QrLoginVerify(qrSession)
		if err != nil {
			fmt.Printf("Error verifying QR login: %v\n", err)
			return
		}

		if !result.Success {
			fmt.Printf("QR login failed: %s\n", result.Message)
			return
		}

		// 保存会话
		session := loginable.SaveSession()
		internal.SetSession(providerName, session)

		fmt.Println("Login successful!")
		fmt.Println("Session:", session)
	},
}

func init() {
	CmdQrlogin.AddCommand(getqrcodeCmd)
	CmdQrlogin.AddCommand(verifyCmd)
}
