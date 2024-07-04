package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"disruptive/cmd/common"
	"disruptive/console/llm"
)

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "file commands",
	Long:  "File commands.",
}

var listFilesCmd = &cobra.Command{
	Use:   "list",
	Short: "list files command",
	Long:  "List files command.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := llm.ListFiles(cmd.Root().Context()); err != nil {
			os.Exit(1)
		}
	},
}

var uploadFilesCmd = &cobra.Command{
	Use:   "upload [filePath]",
	Short: "upload file command",
	Long:  "Upload file command.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]
		if !common.FileExistsValidator(filename) {
			return fmt.Errorf("invalid file: %q", filename)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		purpose, _ := cmd.Flags().GetString("purpose")

		if err := llm.UploadFile(cmd.Root().Context(), filePath, purpose); err != nil {
			os.Exit(1)
		}
	},
}

var deleteFilesCmd = &cobra.Command{
	Use:   "delete [file id]",
	Short: "delete file command",
	Long:  "Delete file command.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileID := args[0]
		if err := llm.DeleteFile(cmd.Root().Context(), fileID); err != nil {
			os.Exit(1)
		}
	},
}

var convertCSVToJSONLCmd = &cobra.Command{
	Use:   "convert [csv filename]",
	Short: "convert csv to jsonl format",
	Long:  "Convert csv to jsonl format.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]
		if !common.FileExistsValidator(filename) {
			return fmt.Errorf("invalid file: %q", filename)
		}
		if !strings.HasSuffix(filename, ".csv") {
			return errors.New("invalid file format")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		if err := llm.ConvertCSVToJSONL(cmd.Root().Context(), filename); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(filesCmd)

	filesCmd.AddCommand(listFilesCmd)

	filesCmd.AddCommand(uploadFilesCmd)
	uploadFilesCmd.Flags().StringP("purpose", "p", "fine-tune", "purpose of the file")

	filesCmd.AddCommand(deleteFilesCmd)

	filesCmd.AddCommand(convertCSVToJSONLCmd)
}
