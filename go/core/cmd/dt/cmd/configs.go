package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"disruptive/cmd/common"
	"disruptive/console/configs"
)

var configsCmd = &cobra.Command{
	Use:   "configs",
	Short: "configs commands",
	Long:  "configs commands.",
}

var configsPersonalitiesCmd = &cobra.Command{
	Use:   "personalities",
	Short: "personality commands",
	Long:  "personality commands.",
}

var addConfigsPersonalitiesCmd = &cobra.Command{
	Use:   "put [system-prompts.csv]",
	Short: "put",
	Long:  "Put.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !common.FileExistsValidator(args[0]) {
			return errors.New("invalid system-prompts.csv file")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := configs.PutPersonalities(cmd.Root().Context(), args[0]); err != nil {
			os.Exit(1)
		}
	},
}

var exportConfigsCmd = &cobra.Command{
	Use:   "export [config_name]",
	Short: "export commands",
	Long:  "Export configs commands.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configName := args[0]

		if err := configs.Export(cmd.Root().Context(), configName); err != nil {
			os.Exit(1)
		}
	},
}

var importConfigsCmd = &cobra.Command{
	Use:   "import [config_name]",
	Short: "import commands",
	Long:  "Import config commands.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configName := args[0]

		all, err := cmd.Flags().GetBool("localize-all")
		if err != nil {
			os.Exit(400)
		}

		if all {
			if err := configs.ImportAllLocalizations(cmd.Root().Context(), configName); err != nil {
				os.Exit(1)
			}
		} else {
			if err := configs.Import(cmd.Root().Context(), configName); err != nil {
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(configsCmd)

	configsCmd.AddCommand(configsPersonalitiesCmd)
	configsPersonalitiesCmd.AddCommand(addConfigsPersonalitiesCmd)

	configsCmd.AddCommand(exportConfigsCmd)

	configsCmd.AddCommand(importConfigsCmd)
	importConfigsCmd.Flags().Bool("localize-all", false, "import all localize files")
}
