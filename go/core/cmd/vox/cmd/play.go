package cmd

import (
	"errors"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"disruptive/cmd/common"
	"disruptive/console/vox/play"
)

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "play commands",
	Long: `Play commands.
	
Characters: 
	alice:		Alice in Wonderland (coqui)
	alice-bella:	Alice in Wonderland
	johnny5:	Johnny Five from Short Circuit
	batman: 	Batman
	sagan:		Carl Sagan
	dora:		Dora the Explorer
	optimus:	Optimus Prime
	spongebob:	SpongeBob SquarePants
	2xl:		2-XL Robot (coqui)
	2xl-johnny5	2-XL Robot


Microphone recording silence effect values:
Example: "1,0.1,3%,1,2.0,3%"
  	"1":	The first parameter after the "silence" effect specifies the number of consecutive audio
		periods (chunks) that must contain audio levels above the threshold for the effect to be applied.
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		words, _ := cmd.Flags().GetInt("words")
		if words < 1 || words > 500 {
			return errors.New("number of words must be in the range 1 to 500")
		}

		creativity, _ := cmd.Flags().GetInt("creativity")
		if creativity < 0 || creativity > 100 {
			return errors.New("creativity must be in the range 0 to 100")
		}

		tokens, _ := cmd.Flags().GetInt("tokens")
		if tokens < 1 || tokens > 1000 {
			return errors.New("tokens must be in the range 1 to 1000")
		}

		questions, _ := cmd.Flags().GetInt("questions")
		if questions < 1 || questions > 10 {
			return errors.New("questions must be in the range 1 to 10")
		}

		session, _ := cmd.Flags().GetInt("session")
		if session < 0 {
			return errors.New("session must be 0 or greater")
		}

		timeout, _ := cmd.Flags().GetInt("timeout")
		if timeout < 0 {
			return errors.New("timeout must be 0 or greater")
		}

		voiceStability, _ := cmd.Flags().GetInt("voice-stability")
		if voiceStability < 0 || voiceStability > 100 {
			return errors.New("voice stability must be in the range 0 to 100")
		}

		voiceSimilarityBoost, _ := cmd.Flags().GetInt("voice-similarity-boost")
		if voiceSimilarityBoost < 0 || voiceSimilarityBoost > 100 {
			return errors.New("voice similarity boost must be in the range 0 to 100")
		}

		language, _ := cmd.Flags().GetString("language")
		if language != "" && len(language) != 2 {
			return errors.New("invalid language")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		age, _ := cmd.Flags().GetString("age")
		character, _ := cmd.Flags().GetString("character")
		creativity, _ := cmd.Flags().GetInt("creativity")
		disableConvo, _ := cmd.Flags().GetBool("disable-convo")
		language, _ := cmd.Flags().GetString("language")
		moderate, _ := cmd.Flags().GetBool("moderate")
		mute, _ := cmd.Flags().GetBool("mute")
		questions, _ := cmd.Flags().GetInt("questions")
		session, _ := cmd.Flags().GetInt("session")
		silence, _ := cmd.Flags().GetString("record-silence")
		timeout, _ := cmd.Flags().GetInt("timeout")
		tokens, _ := cmd.Flags().GetInt("tokens")
		verbose, _ := cmd.Flags().GetBool("verbose")
		voiceStability, _ := cmd.Flags().GetInt("voice-stability")
		voiceSimilarityBoost, _ := cmd.Flags().GetInt("voice-similarity-boost")
		words, _ := cmd.Flags().GetInt("words")

		req := &play.Request{
			Age:                  age,
			Character:            character,
			Creativity:           creativity,
			DisableConvo:         disableConvo,
			Language:             language,
			Moderate:             moderate,
			Mute:                 mute,
			Questions:            questions,
			SessionMemory:        session,
			Silence:              silence,
			TimeoutSeconds:       timeout,
			Tokens:               tokens,
			Verbose:              verbose,
			VoiceStability:       voiceStability,
			VoiceSimilarityBoost: voiceSimilarityBoost,
			Words:                words,
		}

		if !verbose {
			common.SetLogging("error")
		}

		if err := play.Main(cmd.Root().Context(), req); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(playCmd)
	playCmd.Flags().StringP("age", "a", "an adult", "age of the user")
	playCmd.Flags().StringP("character", "c", "sagan", "character to play")
	playCmd.Flags().IntP("creativity", "p", 35, "creativity percentage 0 to 100")
	playCmd.Flags().Bool("disable-convo", false, "disable convo mode which enables next prompting")
	playCmd.Flags().StringP("language", "l", "", "Prefer language using ISO-386-1 format (default multi-language)")
	playCmd.Flags().Bool("moderate", false, "enable moderation")
	playCmd.Flags().BoolP("mute", "m", false, "mute text-to-speech")
	playCmd.Flags().IntP("questions", "q", 4, "number of questions to ask")
	playCmd.Flags().IntP("session", "s", 4, "number of previous user/assistant pairs to use")
	playCmd.Flags().Int("timeout", 10, "seconds allowed between asking the question and the start of the answer")
	playCmd.Flags().IntP("tokens", "t", 250, "max number of tokens")
	playCmd.Flags().Int("voice-stability", 75, "voice stability percentage 0 to 100")
	playCmd.Flags().Int("voice-similarity-boost", 75, "voice similarity boost 0 to 100")
	playCmd.Flags().IntP("words", "n", 50, "number of words in the response")

	switch runtime.GOOS {
	case "linux":
		playCmd.Flags().String("record-silence", "1,0.1,3%,1,1.0,3%", "microphone record silence values")
	case "windows", "darwin":
		playCmd.Flags().String("record-silence", "1,0.1,1%,1,0.3,1%", "microphone record silence values")
	}
}
