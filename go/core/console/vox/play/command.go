package play

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

var (
	alphanumeric = regexp.MustCompile(`[^a-z0-9 ]+`)
)

func processCommand(ctx context.Context, logCtx *slog.Logger, req *Request, question string) (bool, error) {
	c, ok := characters[req.Character]
	if !ok {
		logCtx.Error("invalid character")
		return false, errors.New("invalid character")
	}

	question = strings.ToLower(question)
	question = alphanumeric.ReplaceAllString(question, "")

	parts := strings.Fields(question)

	if len(parts) < 3 {
		return false, nil
	}

	if parts[1] != c.Wake {
		if req.Verbose {
			logCtx.Info("Wake word", "wake", c.Wake, "comparing", parts[1], "found", false)
		}

		return false, nil
	}

	if req.Verbose {
		logCtx.Info("Wake word", "wake", c.Wake, "comparing", parts[1], "found", true)
	}

	if parts[2] == "time" ||
		(len(parts) >= 6) && parts[2] == "what" && parts[3] == "time" && parts[4] == "is" && parts[5] == "it" {
		text := time.Now().Format("The time is Monday, January 06 at 3:04 PM")
		if err := say(ctx, logCtx, req, text); err != nil {
			logCtx.Error("unable to say", "error", err, "character", req.Character)
			return true, err
		}
		return true, nil
	}

	if parts[2] == "birthday" ||
		(len(parts) >= 5) && parts[2] == "its" && parts[3] == "my" && parts[4] == "birthday" {
		if err := playMp3(ctx, logCtx, "birthday.mp3"); err != nil {
			return true, err
		}
		return true, nil
	}

	return false, nil
}
