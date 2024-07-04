package cmd

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"disruptive/cmd/common"
	"disruptive/console/vox/stt"
	c "disruptive/lib/common"
)

var sttCmd = &cobra.Command{
	Use:   "stt",
	Short: "speech-to-text commands",
	Long:  "Speech-to-text commands.",
}

var fileSttCmd = &cobra.Command{
	Use:   "file [audio file]",
	Short: "speech-to-text from an existing audio file (mp3, mp4, mpeg, mpga, m4a, wav, webm)",
	Long:  "speech-to-text from an existing audio file (mp3, mp4, mpeg, mpga, m4a, wav, webm).",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		file := args[0]
		engine, _ := cmd.Flags().GetString("engine")

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

		if engine != "openai" && engine != "deepgram" {
			return errors.New("invalid engine")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]
		timeout, _ := cmd.Flags().GetDuration("timeout")
		language, _ := cmd.Flags().GetString("language")
		engine, _ := cmd.Flags().GetString("engine")

		ctx := context.WithValue(cmd.Root().Context(), c.TimeoutKey, timeout)

		if err := stt.File(ctx, file, language, engine); err != nil {
			os.Exit(1)
		}
	},
}

var recordSttCmd = &cobra.Command{
	Use:   "record [file]",
	Short: "speech-to-text using the microphone",
	Long: `Speech-to-text using the microphone.
	
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
		language, _ := cmd.Flags().GetString("language")
		silence, _ := cmd.Flags().GetString("record-silence")

		if err := stt.Record(cmd.Root().Context(), language, silence); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(sttCmd)

	sttCmd.AddCommand(fileSttCmd)
	fileSttCmd.Flags().StringP("language", "l", "", "Transcribe using to a language using ISO-386-1 format (default multi-language, auto-detect)")
	fileSttCmd.Flags().StringP("engine", "e", "deepgram", "STT engine (openai | deepgram)")

	sttCmd.AddCommand(recordSttCmd)
	recordSttCmd.Flags().StringP("language", "l", "", "Transcribe using to a language using ISO-386-1 format (default multi-language)")

	switch runtime.GOOS {
	case "linux":
		recordSttCmd.Flags().String("record-silence", "1,0.1,3%,1,2.0,3%", "microphone record silence values")
	case "windows":
		recordSttCmd.Flags().String("record-silence", "1,0.1,3%,1,2.0,0.1%", "microphone record silence values")
	}
}
