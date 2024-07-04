package play

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"disruptive/lib/common"
	"disruptive/lib/openai"
)

func processModeration(ctx context.Context, logCtx *slog.Logger, text string) error {
	logCtx = logCtx.With("fid", "vox.play.processModeration")

	m, err := openai.PostModeration(ctx, logCtx, text)
	if err != nil {
		logCtx.Error("unable to get moderation", "error", err)
		return err
	}

	if len(m.Results) != 1 {
		logCtx.Error("invalid moderation results")
		return errors.New("invalid moderation results")
	}

	modResultsStr := []string{}
	modResults := make([]any, 0, len(m.Results[0].Categories))
	for k, v := range m.Results[0].Categories {
		if v {
			switch k {
			case "harassment":
				if !m.Results[0].Categories["harassment/threatening"] {
					modResultsStr = append(modResultsStr, k)
				}
			case "harassment/threatening":
				modResultsStr = append(modResultsStr, "threatening harassment")
			case "violence":
				if !m.Results[0].Categories["violence/graphic"] {
					modResultsStr = append(modResultsStr, k)
				}
			case "violence/graphic":
				modResultsStr = append(modResultsStr, "graphic violence")
			case "sexual":
				if !m.Results[0].Categories["sexual/minor"] {
					modResultsStr = append(modResultsStr, k)
				}
			case "sexual/minors":
				modResultsStr = append(modResultsStr, "sexuality involving minors")
			case "hate":
				if !m.Results[0].Categories["hate/threatening"] {
					modResultsStr = append(modResultsStr, k)
				}
			case "hate/threatening":
				modResultsStr = append(modResultsStr, "threatening hate")
			case "self-harm":
				if !m.Results[0].Categories["self-harm/intent"] && !m.Results[0].Categories["self-harm/instruction"] {
					modResultsStr = append(modResultsStr, k)
				}
			case "self-harm/intent":
				modResultsStr = append(modResultsStr, "self-harm intent")
			case "self-harm/instruction":
				modResultsStr = append(modResultsStr, "self-harm instruction")
			default:
				modResultsStr = append(modResultsStr, k)
			}

			modResults = append(modResults, k, v)
		}
	}

	if len(modResults) > 0 {
		logCtx.Warn("moderation", modResults...)
		return common.ErrModeration{Msg: strings.Join(modResultsStr, " and ")}
	}

	return nil
}
