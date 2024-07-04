package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"disruptive/console/moderate"
)

var moderateCmd = &cobra.Command{
	Use:   "moderate",
	Short: "moderate text",
	Long:  "Moderate text.",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		text, _ := cmd.Flags().GetString("text")
		file, _ := cmd.Flags().GetString("file")

		if text == "" && file == "" {
			return errors.New("file or text is required")
		}

		if text != "" && file != "" {
			return errors.New("file or text may not both be specified")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		text, _ := cmd.Flags().GetString("text")
		file, _ := cmd.Flags().GetString("file")

		if err := moderate.ClassifyText(cmd.Root().Context(), text, file); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(moderateCmd)
	moderateCmd.Flags().StringP("file", "f", "", "file to moderate")
	moderateCmd.Flags().StringP("text", "t", "", "text to moderate")
}
