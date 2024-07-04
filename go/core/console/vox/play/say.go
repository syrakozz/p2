package play

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
)

func say(ctx context.Context, logCtx *slog.Logger, req *Request, text string) error {
	logCtx = logCtx.With("fid", "vox.play.say")

	if req.Mute {
		fmt.Println("[MUTE]", text)
		return nil
	}

	ttsFile, err := os.CreateTemp("", "voxtts*.mp3")
	if err != nil {
		logCtx.Error("unable to create tts file", "error", err)
		return err
	}
	ttsFile.Close()

	if err := tts(ctx, logCtx, req, ttsFile, text); err != nil {
		logCtx.Error("unable to say", "error", err, "character", req.Character)
		return err

	}

	if err := playMp3(ctx, logCtx, ttsFile.Name()); err != nil {
		logCtx.Error("unable to play mp3", "error", err)
		return err
	}

	return nil
}

func playMp3(ctx context.Context, logCtx *slog.Logger, filename string) error {
	logCtx = logCtx.With("fid", "vox.play.playMp3", "filename", filename)
	logCtx.Info("MP3 Playing")

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.CommandContext(ctx, "sox", filename, "-d", "-q")
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/C", "sox", filename, "-t", "waveaudio", "-q")
	default:
		return fmt.Errorf("unknown operating system: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		logCtx.Error("unable to play", "error", err)
		return err
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	return nil
}
