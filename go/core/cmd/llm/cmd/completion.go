package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"disruptive/cmd/common"
	"disruptive/console/llm/completion"
)

var completionCmd = &cobra.Command{
	Use:   "completion [set]",
	Short: "completion command",
	Long:  "Completion command.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		config, _ := cmd.Flags().GetString("config")
		if !common.FileExistsValidator(config) {
			return fmt.Errorf("invalid config: %q", config)
		}

		repeat, _ := cmd.Flags().GetInt("repeat")
		repeatFile, _ := cmd.Flags().GetString("repeat-file")
		if repeat > 0 && repeatFile == "" {
			return errors.New("invalid repeat file")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		set := args[0]
		config, _ := cmd.Flags().GetString("config")
		query, _ := cmd.Flags().GetString("query")
		repeat, _ := cmd.Flags().GetInt("repeat")
		repeatFile, _ := cmd.Flags().GetString("repeat-file")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if !verbose {
			common.SetLogging("error")
		}

		if err := completion.Main(cmd.Root().Context(), config, set, query, repeat, repeatFile, verbose); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
	completionCmd.Flags().StringP("config", "c", "completion.json", "config file containing completion sets")
	completionCmd.Flags().StringP("query", "q", "", "query override for config file.  Use `file:myqueries.txt` to specify a file of queries")
	completionCmd.Flags().IntP("repeat", "r", 0, "Number of times to repeat each question and log it")
	completionCmd.Flags().StringP("repeat-file", "o", "repeat.csv", "output file to save the repeated Q/A")
}
