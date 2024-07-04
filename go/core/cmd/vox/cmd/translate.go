package cmd

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"disruptive/cmd/common"
	"disruptive/console/vox/translate"
	c "disruptive/lib/common"
)

var translateCmd = &cobra.Command{
	Use:   "translate",
	Short: "translate commands",
	Long:  "translate commands.",
}

var fileTranslateCmd = &cobra.Command{
	Use:   "file [audio file]",
	Short: "translate from an existing audio file (mp3, mp4, mpeg, mpga, m4a, wav, webm)",
	Long:  "translate from an existing audio file (mp3, mp4, mpeg, mpga, m4a, wav, webm).",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		file := args[0]

		if !common.FileExistsValidator(file) {
			return errors.New("invalid file")
		}

		if !strings.HasSuffix(file, ".mp3") &&
			!strings.HasSuffix(file, ".mp4") &&
			!strings.HasSuffix(file, ".mpeg") &&
			!strings.HasSuffix(file, ".mpga") &&
			!strings.HasSuffix(file, ".m4a") &&
			!strings.HasSuffix(file, ".wav") &&
			!strings.HasSuffix(file, ".webm") {
			return errors.New("invalid audio file")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx := context.WithValue(cmd.Root().Context(), c.TimeoutKey, timeout)

		if err := translate.File(ctx, file); err != nil {
			os.Exit(1)
		}
	},
}

var recordTranslateCmd = &cobra.Command{
	Use:   "record [file]",
	Short: "translate using the microphone",
	Long: `translate using the microphone.
	
Microphone recording silence effect values:
Example: "1,0.1,3%,1,2.0,3%"
  	"1":	The first parameter specifies the number of consecutive audio periods (chunks)
		that must contain audio levels above the threshold for the effect to be applied.
		In this case, it requires at least one period of audio above the threshold.
  	"0.1":	The second parameter represents the duration (in seconds) that the audio levels must be above
		the threshold within the consecutive periods. In this case, it requires a duration of at least 0.1 seconds.
  	"3%":	The third parameter specifies the threshold level that the audio must surpass to be considered non-silent.
		Here, it is set to 3% of the maximum volume.
  	"1":	This parameter defines the number of consecutive audio periods that must fall below the threshold
		for the effect to be applied. In this case, it requires at least one period of quiet audio.
  	"2.0":	The next parameter sets the duration (in seconds) that the audio levels must remain below the threshold
		within the consecutive quiet periods. Here, it requires a duration of at least 2.0 seconds.
	"3%":	The final parameter represents the threshold level for quiet audio. It is set to 3% of the maximum volume.
`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		silence, _ := cmd.Flags().GetString("record-silence")

		if err := translate.Record(cmd.Root().Context(), silence); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(translateCmd)

	translateCmd.AddCommand(fileTranslateCmd)

	translateCmd.AddCommand(recordTranslateCmd)

	switch runtime.GOOS {
	case "linux":
		recordTranslateCmd.Flags().String("record-silence", "1,0.1,3%,1,2.0,3%", "microphone record silence values")
	case "windows":
		recordTranslateCmd.Flags().String("record-silence", "1,0.1,3%,1,2.0,0.1%", "microphone record silence values")
	}
}
