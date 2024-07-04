package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"disruptive/console/llm"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "model commands",
	Long:  "Model commands.",
}

var listModelsCmd = &cobra.Command{
	Use:   "list",
	Short: "list models command",
	Long:  "List models command.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")

		if id != "" {
			if err := llm.ListModel(cmd.Root().Context(), id); err != nil {
				os.Exit(1)
			}
			return
		}

		if err := llm.ListModels(cmd.Root().Context()); err != nil {
			os.Exit(1)
		}

	},
}

var createModelsCmd = &cobra.Command{
	Use:   "create [file_id]",
	Short: "create models command",
	Long:  "Create models command.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		model, _ := cmd.Flags().GetString("model")

		if model != "ada" && model != "babbage" && model != "curie" && model != "davinci" {
			return errors.New("invalid model")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fileID := args[0]
		epochs, _ := cmd.Flags().GetInt("epochs")
		model, _ := cmd.Flags().GetString("model")
		suffix, _ := cmd.Flags().GetString("suffix")

		if err := llm.CreateModel(cmd.Root().Context(), fileID, model, epochs, suffix); err != nil {
			os.Exit(1)
		}
	},
}

var deleteModelsCmd = &cobra.Command{
	Use:   "delete [model_id]",
	Short: "delete model command",
	Long:  "Delete model command.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		modelID := args[0]

		if err := llm.DeleteModel(cmd.Root().Context(), modelID); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(modelsCmd)

	modelsCmd.AddCommand(listModelsCmd)
	listModelsCmd.Flags().StringP("id", "i", "", "displays the details of a model")

	modelsCmd.AddCommand(createModelsCmd)
	createModelsCmd.Flags().IntP("epochs", "e", 4, "number of epochs")
	createModelsCmd.Flags().StringP("model", "m", "davinci", "training model: ada, babbage, curie, davinci")
	createModelsCmd.Flags().StringP("suffix", "s", "", "suffix to use for the model name")

	modelsCmd.AddCommand(deleteModelsCmd)
}
