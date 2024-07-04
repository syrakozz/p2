package cmd

import (
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"disruptive/console/microphone"
)

var recordCmd = &cobra.Command{
	Use:   "record [file]",
	Short: "record using the microphone",
	Long:  "Record using the microphone.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]
		silence, _ := cmd.Flags().GetString("record-silence")

		if err := microphone.Record(cmd.Root().Context(), silence, file); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(recordCmd)

	switch runtime.GOOS {
	case "linux":
		recordCmd.Flags().String("record-silence", "1,0.1,3%,1,2.0,3%", "microphone record silence values")
	case "windows":
		recordCmd.Flags().String("record-silence", "1,0.1,3%,1,2.0,0.1%", "microphone record silence values")
	}

}
