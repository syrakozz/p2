package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"disruptive/console/accounts"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "accounts commands",
	Long:  "Accounts commands.",
}

var getAccountCmd = &cobra.Command{
	Use:   "get [account_id, email]",
	Short: "get account",
	Long:  "Get an account.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		accountID := args[0]

		if err := accounts.Get(cmd.Root().Context(), accountID); err != nil {
			os.Exit(1)
		}
	},
}

var activateAccountCmd = &cobra.Command{
	Use:   "active [account_id]",
	Short: "activate account",
	Long:  "Activate an account.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		accountID := args[0]

		if err := accounts.PatchInactive(cmd.Root().Context(), accountID, false); err != nil {
			os.Exit(1)
		}
	},
}

var deactivateAccountCmd = &cobra.Command{
	Use:   "inactive [account_id]",
	Short: "deactivate account",
	Long:  "Deactivate an account.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		accountID := args[0]

		if err := accounts.PatchInactive(cmd.Root().Context(), accountID, true); err != nil {
			os.Exit(1)
		}
	},
}

var deleteAccountCmd = &cobra.Command{
	Use:   "delete [account_id]",
	Short: "delete account",
	Long:  "Delete an account.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		accountID := args[0]

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			os.Exit(400)
		}

		if err := accounts.Delete(cmd.Root().Context(), accountID, force); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(accountsCmd)

	accountsCmd.AddCommand(getAccountCmd)

	accountsCmd.AddCommand(activateAccountCmd)

	accountsCmd.AddCommand(deactivateAccountCmd)

	accountsCmd.AddCommand(deleteAccountCmd)
	deleteAccountCmd.Flags().Bool("force", false, "force")

}
