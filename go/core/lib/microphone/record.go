// Package microphone uses the microphone to record a wave file.
package microphone

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"strings"
)

// Record using sox to record from the microphone.
//
// Example effect: "silence 1 0.1 3% 1 2.0 3%"
// 1:   The first parameter after the "silence" effect specifies the number of consecutive audio
// periods (chunks) that must contain audio levels above the threshold for the effect to be applied.
// In this case, it requires at least one period of audio above the threshold.
//
// 0.1: The second parameter represents the duration (in seconds) that the audio levels must be above
// the threshold within the consecutive periods. In this case, it requires a duration of at least 0.1 seconds.
//
// 3%:  The third parameter specifies the threshold level that the audio must surpass to be considered non-silent.
// Here, it is set to 3% of the maximum volume.
//
// 1:   This parameter defines the number of consecutive audio periods that must fall below the threshold
// for the effect to be applied. In this case, it requires at least one period of quiet audio.
//
// 2.0: The next parameter sets the duration (in seconds) that the audio levels must remain below the threshold
// within the consecutive quiet periods. Here, it requires a duration of at least 2.0 seconds.
//
// 3%: The final parameter represents the threshold level for quiet audio. It is set to 3% of the maximum volume.
func Record(ctx context.Context, logCtx *slog.Logger, silence, filename string) error {
	logCtx = logCtx.With("fid", "microphone.Record", "filename", filename, "silence", silence)

	var (
		cmd          *exec.Cmd
		silenceParts []string
		args         = make([]string, 0, 16)
	)

	if silence != "" {
		silenceParts = strings.Split(silence, ",")
		if len(silenceParts) != 6 {
			logCtx.Error("invalid silence values", "silence", silenceParts)
			return errors.New("invalid silence values")
		}
	}

	switch runtime.GOOS {
	case "linux":
		// sox -t alsa default filename silence 1 0.1 3% 1 2.0 3% 2>/dev/null

		if silence == "" {
			silenceParts = []string{"1", "0.1", "3%", "1", "2.0", "3%"}
		}

		args = append(args, "-t", "alsa", "default", filename, "silence")
		args = append(args, silenceParts...)
		args = append(args, "2>/dev/null")

		cmd = exec.CommandContext(ctx, "sox", args...)
	case "darwin":
		// sox -t coreaudio default filename silence 1 0.1 3% 1 2.0 3% 2>/dev/null

		if silence == "" {
			silenceParts = []string{"1", "0.1", "3%", "1", "2.0", "3%"}
		}

		args = append(args, "-t", "coreaudio", "default", filename, "silence")
		args = append(args, silenceParts...)
		args = append(args, "2>/dev/null")

		cmd = exec.CommandContext(ctx, "sox", args...)
	case "windows":
		// cmd /C sox -t waveaudio default -t mp3 filename silence 1 0.1 3% 1 2.0 0.1%

		if silence == "" {
			silenceParts = []string{"1", "0.1", "3%", "1", "2.0", "0.1%"}
		}

		args = append(args, "/C", "sox", "-t", "waveaudio", "default", "-t", "mp3", filename, "silence")
		args = append(args, silenceParts...)

		cmd = exec.CommandContext(ctx, "cmd", args...)
	default:
		return fmt.Errorf("unknown operating system: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			logCtx.Error("unable to run arecord", "error", err)
			return err
		}

		logCtx.Info("Recording canceled")
		return context.Canceled
	}

	return nil
}
