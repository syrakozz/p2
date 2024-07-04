package cmd

import (
	"errors"
	"strconv"

	"github.com/spf13/cobra"

	"disruptive/console/erc"
)

var naicsCmd = &cobra.Command{
	Use:   "naics [code]",
	Short: "lookup business establishment by NAICS code",
	Long:  "Lookup business establishment by NAICS code.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if _, err := strconv.Atoi(args[0]); err != nil {
			return errors.New("invalid NAICS code")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		erc.Naics(cmd.Root().Context(), args[0])
	},
}

func init() {
	rootCmd.AddCommand(naicsCmd)
}
