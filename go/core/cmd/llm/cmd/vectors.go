package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"disruptive/cmd/common"
	"disruptive/console/llm"
)

var vectorsCmd = &cobra.Command{
	Use:   "vectors",
	Short: "vectors commands",
	Long:  "Vectors commands.",
}

var upsertVectorsCmd = &cobra.Command{
	Use:   "upsert [namespace] [file]",
	Short: "list models command",
	Long:  "List models command.",
	Args:  cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[1]
		if !common.FileExistsValidator(filePath) {
			return fmt.Errorf("invalid file: %q", filePath)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		filePath := args[1]
		continueFlag, _ := cmd.Flags().GetBool("continue")

		if err := llm.UpsertVectors(cmd.Root().Context(), namespace, filePath, continueFlag); err != nil {
			os.Exit(1)
		}
	},
}

var deleteVectorsCmd = &cobra.Command{
	Use:   "delete [namespace]",
	Short: "delete command",
	Long:  "Delete command.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		ids, _ := cmd.Flags().GetStringArray("ids")
		deleteAll, _ := cmd.Flags().GetBool("deleteAll")

		if len(ids) < 1 && !deleteAll {
			return errors.New("nothing to delete")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		ids, _ := cmd.Flags().GetStringArray("ids")
		deleteAll, _ := cmd.Flags().GetBool("deleteAll")

		if err := llm.DeleteVectors(cmd.Root().Context(), namespace, ids, deleteAll); err != nil {
			os.Exit(1)
		}
	},
}

var statsVectorsCmd = &cobra.Command{
	Use:   "stats",
	Short: "stats command",
	Long:  "Stats command.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := llm.StatsVectors(cmd.Root().Context()); err != nil {
			os.Exit(1)
		}
	},
}

var queryVectorsCmd = &cobra.Command{
	Use:   "query [namespace] [query]",
	Short: "query command",
	Long:  "Query command.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		query := args[1]
		topK, _ := cmd.Flags().GetInt("topK")

		if err := llm.QueryVectors(cmd.Root().Context(), namespace, query, topK); err != nil {
			os.Exit(1)
		}
	},
}

var fetchVectorCmd = &cobra.Command{
	Use:   "fetch [namespace] [id]",
	Short: "fetch command",
	Long:  "Fetch command.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		id := args[1]

		if err := llm.FetchVector(cmd.Root().Context(), namespace, id); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(vectorsCmd)

	vectorsCmd.AddCommand(upsertVectorsCmd)
	upsertVectorsCmd.Flags().Bool("continue", false, "continue using the seen.csv file")

	vectorsCmd.AddCommand(deleteVectorsCmd)
	deleteVectorsCmd.Flags().StringArray("ids", []string{}, "delete vectors by id")
	deleteVectorsCmd.Flags().Bool("deleteAll", false, "delete all vectors in the namespace")

	vectorsCmd.AddCommand(statsVectorsCmd)

	vectorsCmd.AddCommand(queryVectorsCmd)
	queryVectorsCmd.Flags().IntP("topK", "k", 1, "returns the top K closest number of vectors that matches the query")

	vectorsCmd.AddCommand(fetchVectorCmd)
}
