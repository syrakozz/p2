package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"disruptive/cmd/common"
	"disruptive/console/convert"
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "convert commands",
	Long:  "Convert commands.",
}

var convertTextCmd = &cobra.Command{
	Use:   "text [path]",
	Short: "extract text",
	Long: `Extract text.
This command will only work on Linux and requires the following packages on Debian-based systems.

$ sudo apt-get install poppler-utils wv unrtf tidy`,
	Args: cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !common.FileExistsValidator(args[0]) {
			return errors.New("invalid file")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		output, _ := cmd.Flags().GetString("output")
		meta, _ := cmd.Flags().GetBool("meta")

		if err := convert.TextPath(path, output, meta); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.AddCommand(convertTextCmd)
	convertTextCmd.Flags().StringP("output", "o", "", "output path")
	convertTextCmd.Flags().BoolP("meta", "m", false, "print metadata")
}
