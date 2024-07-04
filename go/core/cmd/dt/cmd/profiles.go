package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"disruptive/console/profiles"
)

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "profile commands",
	Long:  "Profile commands.",
}

var getProfileCmd = &cobra.Command{
	Use:   "get [account_id] [profile_id]",
	Short: "get profile",
	Long:  "Get profile.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		accountID := args[0]
		profileID := args[1]

		if err := profiles.Get(cmd.Root().Context(), accountID, profileID); err != nil {
			os.Exit(1)
		}
	},
}

var getProfileIDsCmd = &cobra.Command{
	Use:   "get-ids [account_id]",
	Short: "get all profile ids",
	Long:  "Get all profile IDs.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		accountID := args[0]

		if err := profiles.GetIDs(cmd.Root().Context(), accountID); err != nil {
			os.Exit(1)
		}
	},
}

var activateProfileCmd = &cobra.Command{
	Use:   "active [account_id] [profile_id]",
	Short: "activate profile",
	Long:  "Activate a profile.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		accountID := args[0]
		profileID := args[1]

		if err := profiles.PatchInactive(cmd.Root().Context(), accountID, profileID, false); err != nil {
			os.Exit(1)
		}
	},
}

var deactivateProfileCmd = &cobra.Command{
	Use:   "inactive [account_id] [profile_id]",
	Short: "deactivate profile",
	Long:  "Deactivate a profile.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		accountID := args[0]
		profileID := args[1]

		if err := profiles.PatchInactive(cmd.Root().Context(), accountID, profileID, true); err != nil {
			os.Exit(1)
		}
	},
}

var deleteProfileCmd = &cobra.Command{
	Use:   "delete [account_id] [profile_id]",
	Short: "delete profile",
	Long:  "Delete a profile.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		accountID := args[0]
		profileID := args[1]

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			os.Exit(400)
		}

		if err := profiles.Delete(cmd.Root().Context(), accountID, profileID, force); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(profilesCmd)

	profilesCmd.AddCommand(getProfileIDsCmd)

	profilesCmd.AddCommand(getProfileCmd)

	profilesCmd.AddCommand(activateProfileCmd)

	profilesCmd.AddCommand(deactivateProfileCmd)

	profilesCmd.AddCommand(deleteProfileCmd)
	deleteProfileCmd.Flags().Bool("force", false, "force")

}
