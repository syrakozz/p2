package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"disruptive/console/users"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "users commands",
	Long:  "Users commands.",
}

var addUsersCmd = &cobra.Command{
	Use:   "add [username] [name]",
	Short: "add user",
	Long:  "Add user.",
	Args:  cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !strings.Contains(args[0], "@") {
			return errors.New("invalid username")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		name := args[1]
		permissions, _ := cmd.Flags().GetStringSlice("permissions")

		fmt.Print("Password: ")
		password, err := term.ReadPassword(0)
		fmt.Println()
		if err != nil {
			slog.Error("unable to read password")
			return
		}

		fmt.Print("Verify Password: ")
		password2, err := term.ReadPassword(0)
		fmt.Println()
		if err != nil {
			slog.Error("unable to read password")
			return
		}

		if string(password) != string(password2) {
			slog.Error("passwords don't match")
			return
		}

		if err := users.Add(cmd.Root().Context(), username, name, string(password), permissions); err != nil {
			os.Exit(1)
		}
	},
}

var getUsersCmd = &cobra.Command{
	Use:   "get [username]",
	Short: "get user",
	Long:  "Get user.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !strings.Contains(args[0], "@") {
			return errors.New("invalid username")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if err := users.Get(cmd.Root().Context(), name); err != nil {
			os.Exit(1)
		}
	},
}

var modifyUsersCmd = &cobra.Command{
	Use:   "modify [username]",
	Short: "modify user",
	Long:  "Modify user.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !strings.Contains(args[0], "@") {
			return errors.New("invalid username")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		name, _ := cmd.Flags().GetString("name")
		permissions, _ := cmd.Flags().GetStringSlice("permissions")
		p, _ := cmd.Flags().GetBool("password")

		var (
			password []byte
			err      error
		)

		if p {
			fmt.Print("Password: ")
			password, err = term.ReadPassword(0)
			fmt.Println()
			if err != nil {
				slog.Error("unable to read password")
				return
			}

			fmt.Print("Verify Password: ")
			password2, err := term.ReadPassword(0)
			fmt.Println()
			if err != nil {
				slog.Error("unable to read password")
				return
			}

			if string(password) != string(password2) {
				slog.Error("passwords don't match")
				return
			}
		}

		if err := users.Modify(cmd.Root().Context(), username, name, string(password), permissions); err != nil {
			os.Exit(1)
		}
	},
}

var deleteUsersCmd = &cobra.Command{
	Use:   "delete [username]",
	Short: "delete user",
	Long:  "Delete user.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !strings.Contains(args[0], "@") {
			return errors.New("invalid username")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]

		if err := users.Delete(cmd.Root().Context(), username); err != nil {
			os.Exit(1)
		}
	},
}

var verifyPasswordUsersCmd = &cobra.Command{
	Use:   "verify [username]",
	Short: "verify user",
	Long:  "verify user.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !strings.Contains(args[0], "@") {
			return errors.New("invalid username")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]

		fmt.Print("Password: ")
		password, err := term.ReadPassword(0)
		fmt.Println()
		if err != nil {
			slog.Error("unable to read password")
			os.Exit(1)
		}

		if err := users.VerifyPassword(cmd.Root().Context(), username, string(password)); err != nil {
			os.Exit(1)
		}
	},
}

var loginUsersCmd = &cobra.Command{
	Use:   "login [username]",
	Short: "login user",
	Long:  "login user.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]

		expire, _ := cmd.Flags().GetDuration("expire")

		if err := users.Login(cmd.Root().Context(), username, expire); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(usersCmd)

	usersCmd.AddCommand(addUsersCmd)
	addUsersCmd.Flags().StringSliceP("permissions", "p", []string{}, "permissions")

	usersCmd.AddCommand(getUsersCmd)

	usersCmd.AddCommand(modifyUsersCmd)
	modifyUsersCmd.Flags().StringP("name", "n", "", "name")
	modifyUsersCmd.Flags().BoolP("password", "", false, "set password")
	modifyUsersCmd.Flags().StringSliceP("permissions", "p", []string{}, "permissions")

	usersCmd.AddCommand(deleteUsersCmd)

	usersCmd.AddCommand(verifyPasswordUsersCmd)

	usersCmd.AddCommand(loginUsersCmd)
	loginUsersCmd.Flags().DurationP("expire", "e", 12*time.Hour, "token expiration duration")
}
