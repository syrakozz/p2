package play

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"disruptive/lib/common"
	"disruptive/lib/microphone"
)

type (
	// Request contains input parameters
	Request struct {
		Age                    string
		Character              string
		Words                  int
		Creativity             int
		Moderate               bool
		Mood                   string
		Mute                   bool
		Questions              int
		SessionMemory          int
		TimeoutSeconds         int
		Tokens                 int
		Verbose                bool
		VoiceSimilarityBoost   int
		VoiceStability         int
		VoiceStyleExaggeration int
		Silence                string
		DisableConvo           bool
		Language               string
	}

	playSessionMemoryEntry struct {
		User      string `json:"user"`
		Assistant string `json:"assistant"`
	}

	playSessionMemory struct {
		Entries []playSessionMemoryEntry `json:"entries"`
	}
)

var (
	pLen    int
	errStop = errors.New("Goodbye")
)

// Main runs the Play cycle of playing the audio question/answer multiple times.
func Main(ctx context.Context, req *Request) error {
	logCtx := slog.With("fid", "vox.play.Main")

	if err := readConfig(logCtx); err != nil {
		logCtx.Error("unabel to read config", "error", err)
		return err
	}

	c, ok := characters[req.Character]
	if !ok {
		logCtx.Error("invalid character", "character", req.Character)
		return errors.New("invalid character")
	}

	// Set system level vars for chosen character

	if c.SystemPromptTemplate != "" {
		systemPromptTemplate = c.SystemPromptTemplate
	}

	// Session Memory

	sessionMemory := &playSessionMemory{}

	if b, err := os.ReadFile("session_" + req.Character + ".json"); err == nil {
		if err := json.Unmarshal(b, sessionMemory); err != nil {
			logCtx.Warn("unable unmarshal session memory", "error", err)
			return err
		}
	}

	defer func() {
		b, err := json.MarshalIndent(sessionMemory, "", "  ")
		if err != nil {
			logCtx.Error("unable to marshal session memory", "error", err)
		}

		if err := os.WriteFile("session_"+req.Character+".json", b, 0644); err != nil {
			logCtx.Error("unable to write session memory", "error", err)
		}
	}()

	var text string
	if len(c.Greetings) > 0 {
		text = c.Greetings[rand.Intn(len(c.Greetings))] + "\n"
	} else {
		text = fmt.Sprintf("Hello, I'm %s, what are you curious about today?\n", c.ShortName)
	}

	if err := say(ctx, logCtx, req, text); err != nil {
		logCtx.Error("unable to say", "error", err, "character", req.Character)
		return err
	}

	for i := req.Questions; i > 0; i-- {
		if req.DisableConvo && i != req.Questions {
			text := "What is your next question?"
			if err := say(ctx, logCtx, req, text); err != nil {
				logCtx.Error("unable to say", "error", err, "character", req.Character)
				return err
			}
		}

		if err := play(ctx, logCtx, req, sessionMemory); err != nil {
			if errors.Is(err, common.ErrLimit) ||
				errors.Is(err, common.ErrModeration{}) ||
				errors.Is(err, context.DeadlineExceeded) {
				i++
				continue
			}
			if errors.Is(err, errStop) {
				return nil
			}
			return err
		}
	}

	text = "Have a great day!"
	if err := say(ctx, logCtx, req, text); err != nil {
		logCtx.Error("unable to say", "error", err, "character", req.Character)
		return err
	}

	return nil
}

func play(ctx context.Context, logCtx *slog.Logger, req *Request, sessionMemory *playSessionMemory) error {
	logCtx = logCtx.With("fid", "vox.play.play")
	logCtx.Info("Play Start")

	durs := map[string]time.Duration{}
	aLen := 0

	if req.Verbose {
		defer func() {
			if aLen == 0 || aLen+pLen == 0 {
				stt := durs["stt"]
				moderation := durs["moderation"]

				fmt.Println("")
				fmt.Println("Durations")
				fmt.Println("        Moderation:", moderation.String())
				fmt.Println("               STT:", stt.String())
				fmt.Println()
				return
			}

			stt := durs["stt"]
			moderation := durs["moderation"]
			chat := durs["chat"]
			tts := durs["tts"]
			recordToPlay := durs["recordToPlay"]
			total := stt + moderation + chat + tts

			fmt.Println("")
			fmt.Println("Durations")
			fmt.Println("        Moderation:", moderation.String())
			fmt.Println("               STT:", stt.String())
			fmt.Printf("              Chat: %s (P: %d, A: %d, Chat/PA %s)\n", chat.String(), pLen, aLen, time.Duration(chat.Nanoseconds()/int64(pLen+aLen)).String())
			fmt.Println("               TTS:", tts.String())
			fmt.Println("  STT+Mod+Chat+TTS:", total.String())
			fmt.Println("    Record to play:", recordToPlay.String())
			fmt.Println("              Slop:", recordToPlay-total)
			fmt.Println()
		}()
	}

	// Create temp speech-to-text file
	sttFile, err := os.CreateTemp("", "voxstt*.mp3")
	if err != nil {
		logCtx.Error("unable to create stt temp file", "error", err)
		return err
	}

	sttFilename := sttFile.Name()
	sttFile.Close()

	ttsFile, err := os.CreateTemp("", "voxtts*.mp3")
	if err != nil {
		logCtx.Error("unable to create tts temp file", "error", err)
		return err
	}

	// Create temp text-to-speech file
	ttsFilename := ttsFile.Name()
	ttsFile.Close()

	logCtx.Info("Recording question", "filename", sttFilename)
	if err := microphone.Record(ctx, logCtx, req.Silence, sttFilename); err != nil {
		logCtx.Error("unable to record", "error", err)
		return err
	}
	tEndRecord := time.Now()

	// Convert question from speech to text.

	logCtx.Info("Transcribing question", "filename", sttFilename)
	t := time.Now()
	question, err := stt(ctx, logCtx, sttFilename, req.Language)
	if err != nil {
		logCtx.Error("unable to convert speech to text", "error", err)
		return err
	}
	durs["stt"] = time.Since(t)

	if req.Verbose {
		fmt.Printf("\nQUESTION:\n%s\n\n", question)
	}

	if ok, err := processCommand(ctx, logCtx, req, question); err != nil {
		logCtx.Error("unable to process command", "error", err)
		return err
	} else if ok {
		return nil
	}

	if req.Moderate {
		t := time.Now()
		if err := processModeration(ctx, logCtx, question); err != nil {
			if errors.Is(err, common.ErrModeration{}) {
				text := fmt.Sprintf("The question contains %s which is not allowed.  Please try again!", err.Error())
				if err := say(ctx, logCtx, req, text); err != nil {
					logCtx.Error("unable to say", "error", err, "character", req.Character)
					return err
				}
				durs["moderation"] = time.Since(t)
				return err
			}
		}
		durs["moderation"] = time.Since(t)
	}

	// Answer question.

	err = func() error {
		s := req.TimeoutSeconds
		if s < 1 {
			s = 60
		}

		ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(s)*time.Second)
		defer cancel()

		// text question to text answer

		t := time.Now()

		answer, err := response(ctxWithTimeout, logCtx, question, req, sessionMemory)
		durs["chat"] = time.Since(t)
		aLen = len(answer)

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				text := "I'm sorry, the response took too long. Please try again!"
				if err := say(ctx, logCtx, req, text); err != nil {
					logCtx.Error("unable to say", "error", err, "character", req.Character)
					return err
				}
				logCtx.Warn("answer timeout", "error", err)
				return err
			}
			logCtx.Error("unable to get an answer", "error", err)
			return err
		}

		if req.Verbose {
			fmt.Printf("\nANSWER:\n%s\n\n", answer)
		}

		// Convert answer from text to speech.

		if req.Mute {
			fmt.Println("[MUTE]", answer)
			return nil
		}

		logCtx.Info("Encoding answer", "filename", ttsFilename, "len", len(answer))
		t = time.Now()
		if err := tts(ctxWithTimeout, logCtx, req, ttsFile, answer); err != nil {
			if errors.Is(err, common.ErrLimit) {
				if err := say(ctx, logCtx, req, "The answer is too long, please try again!"); err != nil {
					logCtx.Error("unable to say", "error", err, "character", req.Character)
					return err
				}
				return err
			}
			if errors.Is(err, context.DeadlineExceeded) {
				text := "I'm sorry, the response took too long. Please try again!"
				if err := say(ctx, logCtx, req, text); err != nil {
					logCtx.Error("unable to say", "error", err, "character", req.Character)
					return err
				}
				logCtx.Warn("tts timeout", "error", err)
				return err
			}

			logCtx.Error("unable to convert text to speech", "error", err)
			return err
		}
		durs["tts"] = time.Since(t)

		return nil
	}()

	if err != nil {
		return err
	}

	// Play answer.

	if req.Mute {
		return nil
	}

	durs["recordToPlay"] = time.Since(tEndRecord)
	if err := playMp3(ctx, logCtx, ttsFilename); err != nil {
		logCtx.Error("unable to play mp3", "error", err)
		return err
	}

	if len(question) < 10 && strings.Contains(strings.ToLower(question), "goodbye") {
		return errStop
	}

	return nil
}
