package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"disruptive/console/characters"
)

var charactersCmd = &cobra.Command{
	Use:   "characters",
	Short: "characters commands",
	Long:  "Characters commands.",
}

var exportCharacterCmd = &cobra.Command{
	Use:   "export [character_version] [language]",
	Short: "export commands",
	Long:  "Export character commands.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		characterVersion := args[0]
		language := args[1]

		if err := characters.Export(cmd.Root().Context(), characterVersion, language); err != nil {
			os.Exit(1)
		}
	},
}

var importCharacterCmd = &cobra.Command{
	Use:   "import [character_version] [language]",
	Short: "import commands",
	Long:  "Import character commands.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		characterVersion := args[0]
		language := args[1]

		if language == "all" {
			if err := characters.ImportAll(cmd.Root().Context(), characterVersion); err != nil {
				os.Exit(1)
			}
		} else {
			if err := characters.Import(cmd.Root().Context(), characterVersion, language); err != nil {
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(charactersCmd)

	charactersCmd.AddCommand(exportCharacterCmd)

	charactersCmd.AddCommand(importCharacterCmd)
}
