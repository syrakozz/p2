package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"disruptive/console/llm/stress"
)

var stressCmd = &cobra.Command{
	Use:   "stress",
	Short: "stress commands",
	Long:  "Stress commands.",
}

var embeddingsStressCmd = &cobra.Command{
	Use:   "embeddings",
	Short: "embeddings stress command",
	Long:  "Embeddings stress command.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		num, _ := cmd.Flags().GetInt("num")
		text, _ := cmd.Flags().GetString("text")

		if err := stress.Embeddings(cmd.Root().Context(), text, num); err != nil {
			os.Exit(1)
		}
	},
}

var whisperStressCmd = &cobra.Command{
	Use:   "whisper [file]",
	Short: "whisper stress command",
	Long:  "Whisper stress command.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]
		num, _ := cmd.Flags().GetInt("num")

		if err := stress.Whisper(cmd.Root().Context(), file, num); err != nil {
			os.Exit(1)
		}
	},
}

var elevenlabsStressCmd = &cobra.Command{
	Use:   "elevenlabs [text]",
	Short: "Elevenlabs stress command",
	Long:  "Elevenlabs stress command.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text := args[0]
		num, _ := cmd.Flags().GetInt("num")

		if err := stress.Elevenlabs(cmd.Root().Context(), text, num); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(stressCmd)

	stressCmd.AddCommand(embeddingsStressCmd)
	embeddingsStressCmd.Flags().IntP("num", "n", 1, "number of iterations")
	embeddingsStressCmd.Flags().StringP("text", "t", "", "text")

	stressCmd.AddCommand(whisperStressCmd)
	whisperStressCmd.Flags().IntP("num", "n", 1, "number of iterations")

	stressCmd.AddCommand(elevenlabsStressCmd)
	elevenlabsStressCmd.Flags().IntP("num", "n", 1, "number of iterations")
}
