package play

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/configs"
	"disruptive/lib/openai"
	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/characters"
	"disruptive/pkg/vox/moderate"
	"disruptive/pkg/vox/profiles"
)

// Result contains TTT response data.
type Result struct {
	Moderation     *moderate.Response `json:"moderation,omitempty"`
	Predefined     bool               `json:"predefined,omitempty"`
	Response       string             `json:"response"`
	SessionID      int                `json:"session_id"`
	TokensPrompt   int                `json:"tokens_prompt"`
	TokensResponse int                `json:"tokens_response"`
}

// TTT runs one LLM prompt/response.
func TTT(ctx context.Context, logCtx *slog.Logger, profile *profiles.Document, characterVersion, tttModel, userPrompt, audioID string) (*Result, error) {
	fid := slog.String("fid", "vox.characters.play.TTT")
	t := time.Now()

	account := ctx.Value(common.AccountKey).(accounts.Document)

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	c, err := characters.GetCharacter(ctx, logCtx, characterVersion, profileCharacter.Language)
	if err != nil {
		logCtx.Error("character doc not found", fid, "error", err)
		return nil, err
	}

	session, err := characters.GetSession(ctx, logCtx, profile.ID, c.Character)
	if err != nil {
		logCtx.Error("unable to get session", fid, "error", err)
		return nil, err
	}

	if session.LastUserAudio[audioID].Mode != "" {
		profileCharacter.Mode = session.LastUserAudio[audioID].Mode
	}

	mode, ok := c.Modes[profileCharacter.Mode]
	if !ok {
		profileCharacter.Mode = "conversation"
		mode, ok = c.Modes[profileCharacter.Mode]
		if !ok {
			logCtx.Error("invalid mode", fid)
			return nil, common.ErrNotFound{Msg: "invalid mode"}
		}
	}

	profile.Characters[characterName] = profileCharacter

	if account.DeveloperMode {
		if creativity, ok := account.DeveloperModeMap["character_creativity"].(int); ok {
			mode.Creativity = creativity
		}

		if entries, ok := account.DeveloperModeMap["character_session_entries"].(int); ok {
			mode.SessionEntries = entries
		}
	}

	language := profileCharacter.Language
	if language == "" && audioID != "" && !session.LastUserAudio[audioID].Predefined {
		language = session.LastUserAudio[audioID].DetectedLanguage
	}
	logCtx.Info("stt language", "region", config.VARS.STTRegion, "language", language)

	localize, err := configs.GetLocalization(ctx, logCtx, strings.Split(characterVersion, "_")[1], language)
	if err != nil {
		logCtx.Error("unable to get character localize configs", fid, "error", err)
		return nil, err
	}

	var (
		chatReq      *openai.ChatRequest
		numTokens    int
		promptCost   float64
		responseCost float64
	)

	if tttModel != "" {
		c.Model = tttModel
	}

	if c.Model == "gpt-4-turbo" {
		c.Model = characters.GPT4Turbo
	}

	switch c.Model {
	case "gpt-3.5-turbo", characters.GPT4Turbo:
		switch profileCharacter.Mode {
		case "conversation", "story", "teach_me_something":
			switch c.Version {
			case 1:
				chatReq, numTokens = conversationGPTPromptBuilderV1(profile, &c, &mode, &session, userPrompt, &localize, session.LastUserAudio[audioID].Predefined)
			case 2:
				chatReq, numTokens = conversationGPTPromptBuilderV2(profile, &c, &mode, &session, userPrompt, &localize, session.LastUserAudio[audioID].Predefined)
			default:
				logCtx.Error("unable to build conversation/story prompt invalid version", fid, "mode", profileCharacter.Mode, "version", c.Version)
				return nil, errors.New("unable to build prompt invalid version")
			}

		case "fun":
			c.Model = characters.GPT4Turbo

			switch c.Version {
			case 1:
				chatReq, numTokens = funGPTPromptBuilderV1(profile, &c, &mode, &session, userPrompt, &localize)
			case 2:
				chatReq, numTokens = funGPTPromptBuilderV2(profile, &c, &mode, &session, userPrompt, &localize)
			default:
				logCtx.Error("unable to build fun prompt invalid version", fid, "mode", profileCharacter.Mode, "version", c.Version)
				return nil, errors.New("unable to build prompt invalid version")
			}

		default:
			logCtx.Error("unable to build prompt invalid mode", fid, "mode", profileCharacter.Mode)
			return nil, errors.New("unable to build prompt invalid mode")
		}

		if chatReq.Model == "gpt-3.5-turbo" && numTokens > characters.MaxPromptTokens35Turbo {
			logCtx.Info("gpt-3-turbo token length exceeded. switching to gpt-4-turbo model.", fid, "prompt_tokens", numTokens, "model", chatReq.Model)
			chatReq.Model = characters.GPT4Turbo
		}

		if chatReq.Model == characters.GPT4Turbo && numTokens > characters.MaxPromptTokensGPT4Turbo {
			logCtx.Error("gpt-4-turbo token length exceeded.", fid, "prompt_tokens", numTokens, "model", chatReq.Model)
			return nil, common.ErrBadRequest{Msg: "Request exceeds max token length."}
		}

		res, err := playGPT(ctx, logCtx, profile, &c, &session, chatReq, userPrompt, audioID)
		if err != nil {
			logCtx.Error("unable to post chat to GPT", fid)
			return nil, errors.New("unable to post chat to GPT")
		}

		switch chatReq.Model {
		case characters.GPT4Turbo:
			promptCost = float64(res.TokensPrompt) * config.VARS.GPT4TurboPromptCost / 1000.0
			responseCost = float64(res.TokensResponse) * config.VARS.GPT4TurboResponseCost / 1000.0
		default:
			promptCost = float64(res.TokensPrompt) * config.VARS.GPT35TurboPromptCost / 1000.0
			responseCost = float64(res.TokensResponse) * config.VARS.GPT35TurboResponseCost / 1000.0
		}

		logCtx.Info(
			"cost",
			"feature", "chat",
			"predefined", res.Predefined,
			"mode", profileCharacter.Mode,
			"model", c.Model,
			"version", c.Version,
			"tokens_prompt", res.TokensPrompt,
			"tokens_response", res.TokensResponse,
			"tokens_total", res.TokensPrompt+res.TokensResponse,
			"cost_prompt", fmt.Sprintf("%.7f", promptCost),
			"cost_response", fmt.Sprintf("%.7f", responseCost),
			"cost_total", fmt.Sprintf("%.7f", promptCost+responseCost),
		)

		logCtx.Info("duration", "duration", time.Since(t).Milliseconds(), "span", "ttt")

		return res, nil
	}

	logCtx.Error("invalid model", fid, "model", c.Model)
	return nil, errors.New("invalid model")
}
