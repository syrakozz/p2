package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"disruptive/cmd/common"
	"disruptive/console/memory"
)

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "memory commands",
	Long:  "memory commands.",
}

var addMemoryCmd = &cobra.Command{
	Use:   "add",
	Short: "add commands",
	Long:  "add commands.",
}

var mboxAddMemoryCmd = &cobra.Command{
	Use:   "mbox [namespace] [path]",
	Short: "mbox text",
	Long:  "mbox text.",
	Args:  cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !common.FileExistsValidator(args[1]) {
			return errors.New("invalid file")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		path := args[1]

		console, _ := cmd.Flags().GetBool("console")
		pinecone, _ := cmd.Flags().GetBool("pinecone")
		continueFlag, _ := cmd.Flags().GetBool("continue")

		if pinecone {
			if err := memory.ProcessMboxPinecone(cmd.Root().Context(), path, namespace, continueFlag); err != nil {
				os.Exit(1)
			}
		}

		if console {
			if err := memory.ProcessMboxConsole(cmd.Root().Context(), path, namespace, continueFlag); err != nil {
				os.Exit(1)
			}
		}

	},
}

var deleteMemoryCmd = &cobra.Command{
	Use:   "delete",
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

		if err := memory.Delete(cmd.Root().Context(), namespace, ids, deleteAll); err != nil {
			os.Exit(1)
		}
	},
}

var queryMemoryCmd = &cobra.Command{
	Use:   "query [namespace] [query]",
	Short: "query command",
	Long:  "Query command.",
	Args:  cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		query := args[1]
		if len(query) > 10000 {
			return errors.New("query length too long")
		}

		if topK, _ := cmd.Flags().GetInt("topK"); topK < 1 {
			return errors.New("invalid topK")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		query := args[1]
		topK, _ := cmd.Flags().GetInt("topK")

		if err := memory.Query(cmd.Root().Context(), namespace, query, topK); err != nil {
			os.Exit(1)
		}
	},
}

var statsMemoryCmd = &cobra.Command{
	Use:   "stats",
	Short: "stats command",
	Long:  "Stats command.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := memory.Stats(cmd.Root().Context()); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(memoryCmd)

	memoryCmd.AddCommand(addMemoryCmd)

	addMemoryCmd.AddCommand(mboxAddMemoryCmd)
	mboxAddMemoryCmd.Flags().BoolP("console", "", false, "ingest to console")
	mboxAddMemoryCmd.Flags().BoolP("pinecone", "", false, "ingest to pinecone")
	mboxAddMemoryCmd.Flags().BoolP("continue", "", false, "continue using the associated mbox.csv file")

	memoryCmd.AddCommand(deleteMemoryCmd)
	deleteMemoryCmd.Flags().StringArrayP("ids", "", []string{}, "delete memory document by id")
	deleteMemoryCmd.Flags().BoolP("deleteAll", "", false, "delete all documents in the namespace")

	memoryCmd.AddCommand(queryMemoryCmd)
	queryMemoryCmd.Flags().IntP("topK", "k", 4, "top K search results to use in the answer")

	memoryCmd.AddCommand(statsMemoryCmd)
}
