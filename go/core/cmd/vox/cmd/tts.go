package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"disruptive/console/tts"
)

var ttsCmd = &cobra.Command{
	Use:   "tts",
	Short: "text-to-speech commands",
	Long:  "Text-to-speech commands.",
}

var elevenlabsTtsCmd = &cobra.Command{
	Use:   "elevenlabs",
	Short: "elevenlabs tts commands",
	Long: `Elevenlabs tts commands.
	
Voices:	
    Female:     rachel, domi, bella, elli
    Male:       antoni, josh, arnold, adam, sam
    Characters: optimus, batman, spongebob
	`,
	Args: cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		text, _ := cmd.Flags().GetString("text")
		file, _ := cmd.Flags().GetString("file")

		if text == "" && file == "" {
			return errors.New("file or text is required")
		}

		if text != "" && file != "" {
			return errors.New("file or text may not both be specified")
		}

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		text, _ := cmd.Flags().GetString("text")
		file, _ := cmd.Flags().GetString("file")
		output, _ := cmd.Flags().GetString("output")
		voice, _ := cmd.Flags().GetString("voice")
		language, _ := cmd.Flags().GetString("language")

		if err := tts.Elevenlabs(cmd.Root().Context(), file, text, output, voice, language); err != nil {
			os.Exit(1)
		}
	},
}

var coquiTtsCmd = &cobra.Command{
	Use:   "coqui",
	Short: "coqui tts commands",
	Long:  `Coqui tts commands.`,
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		text, _ := cmd.Flags().GetString("text")
		file, _ := cmd.Flags().GetString("file")
		voice, _ := cmd.Flags().GetString("voice")
		prompt, _ := cmd.Flags().GetString("prompt")

		if text == "" && file == "" {
			return errors.New("file or text is required")
		}

		if text != "" && file != "" {
			return errors.New("file or text may not both be specified")
		}

		if voice == "" && prompt == "" {
			return errors.New("voice or prompt is required")
		}

		if text != "" && file != "" {
			return errors.New("voice or prompt may not both be specified")
		}

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		text, _ := cmd.Flags().GetString("text")
		file, _ := cmd.Flags().GetString("file")
		output, _ := cmd.Flags().GetString("output")
		voice, _ := cmd.Flags().GetString("voice")
		prompt, _ := cmd.Flags().GetString("prompt")

		if err := tts.Coqui(cmd.Root().Context(), file, text, output, prompt, voice); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(ttsCmd)

	ttsCmd.AddCommand(elevenlabsTtsCmd)
	elevenlabsTtsCmd.Flags().StringP("file", "f", "", "file to convert to speech")
	elevenlabsTtsCmd.Flags().StringP("text", "t", "", "text to convert to speech")
	elevenlabsTtsCmd.Flags().StringP("output", "o", "audio.mp3", "output file to save audio. Set to '-' for stdout streaming")
	elevenlabsTtsCmd.Flags().StringP("voice", "v", "rachel", "audio voice to use")
	elevenlabsTtsCmd.Flags().StringP("language", "l", "en-US", "language to use")

	ttsCmd.AddCommand(coquiTtsCmd)
	coquiTtsCmd.Flags().StringP("file", "f", "", "file to convert to speech")
	coquiTtsCmd.Flags().StringP("text", "t", "", "text to convert to speech")
	coquiTtsCmd.Flags().StringP("output", "o", "audio.wav", "output file to save audio")
	coquiTtsCmd.Flags().StringP("voice", "v", "", "audio voice to use")
	coquiTtsCmd.Flags().StringP("prompt", "p", "", "voice prompt")

}
