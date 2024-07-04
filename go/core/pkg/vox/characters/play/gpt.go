package play

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"regexp"
	"slices"
	"strings"
	"time"

	"disruptive/lib/configs"
	"disruptive/lib/openai"
	"disruptive/pkg/vox/characters"
	"disruptive/pkg/vox/profiles"

	"github.com/pkoukk/tiktoken-go"
)

func playGPT(ctx context.Context, logCtx *slog.Logger, profile *profiles.Document, character *characters.Character,
	session *characters.SessionDocument, chatReq *openai.ChatRequest, userPrompt, audioID string) (*Result, error) {
	fid := slog.String("fid", "vox.characters.play.playGPT")

	t := time.Now()

	chatRes, err := openai.PostChat(ctx, logCtx, *chatReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logCtx.Error("timeout", fid, "error", err)
			return nil, err
		}
		logCtx.Error("unable to get chat response", fid, "error", err)
		return nil, err
	}

	// remove words (dont_say)
	text := chatRes.Text
	for _, v := range profile.DontSay {
		r := `(?i)\b` + regexp.QuoteMeta(v) + `\b`
		regex, err := regexp.Compile(r)
		if err != nil {
			logCtx.Warn("unable to compile regex", fid, "error", err, "regex", r)
			continue
		}

		text = regex.ReplaceAllLiteralString(text, "")
	}

	// replace words
	for k, v := range profile.ReplaceWords {
		for i, w := range v {
			profile.ReplaceWords[k][i] = regexp.QuoteMeta(w)
		}

		r := `(?i)\b(` + strings.Join(profile.ReplaceWords[k], "|") + `)\b`
		regex, err := regexp.Compile(r)
		if err != nil {
			logCtx.Warn("unable to compile regex", fid, "error", err, "regex", r)
			continue
		}

		if k == "_" {
			k = ""
		}

		text = regex.ReplaceAllLiteralString(text, k)
	}

	text = strings.ReplaceAll(text, "  ", " ")

	sessionEntry := characters.SessionEntry{
		Assistant:      text,
		Mode:           profile.Characters[character.Character].Mode,
		Timestamp:      t,
		TokensPrompt:   chatRes.UsagePrompt,
		TokensResponse: chatRes.UsageResponse,
		User:           userPrompt,
	}

	sessionID, err := characters.AddSessionEntry(ctx, logCtx, profile.ID, character.Character, audioID, *session, sessionEntry)
	if err != nil {
		logCtx.Warn("unable to add session entry", fid, "error", err)
	}

	if sessionID == 0 {
		logCtx.Warn("unable to get sessionID", fid)
	}

	p := &Result{
		Response:       text,
		Predefined:     session.LastUserAudio[audioID].Predefined,
		SessionID:      sessionID,
		TokensPrompt:   chatRes.UsagePrompt,
		TokensResponse: chatRes.UsageResponse,
	}

	return p, nil
}

func conversationGPTPromptBuilderV1(profile *profiles.Document, character *characters.Character, mode *characters.Mode,
	session *characters.SessionDocument, userPrompt string, localize *configs.Localize, predefined bool) (*openai.ChatRequest, int) {
	systemPrompt := strings.Builder{}
	systemPrompt.WriteString(mode.CharacterPrompt)

	if len(character.TraitsPositive) > 0 {
		systemPrompt.WriteString(localize.Character["traits_positive"])
		systemPrompt.WriteString(strings.Join(character.TraitsPositive, ", "))
		systemPrompt.WriteString(".")
	}

	if len(character.TraitsNegative) > 0 {
		systemPrompt.WriteString(localize.Character["traits_negative"])
		systemPrompt.WriteString(strings.Join(character.TraitsNegative, ", "))
		systemPrompt.WriteString(".")
	}

	systemPrompt.WriteString(localize.Character["guardrail_prompt"])

	systemPromptString := systemPrompt.String()
	chatReq := openai.ChatRequest{
		Model:      character.Model,
		Messages:   []openai.ChatMessage{{Role: "system", Content: systemPromptString}},
		Creativity: mode.Creativity,
	}

	if !predefined {
		entryNumber := len(session.Entries) + session.StartEntry
		numEntries := 0
		memoryMessages := []openai.ChatMessage{}
		for numEntries <= mode.SessionEntries && entryNumber >= session.StartEntry {
			iStr := fmt.Sprintf("%06d", entryNumber)

			if session.Entries[iStr].EndSequence {
				break
			}

			if session.Entries[iStr].Moderation != nil && !session.Entries[iStr].Moderation.Triggered {
				memoryMessages = append(memoryMessages, openai.ChatMessage{Role: "assistant", Content: session.Entries[iStr].Assistant})
				memoryMessages = append(memoryMessages, openai.ChatMessage{Role: "user", Content: session.Entries[iStr].User})
				numEntries++
			}

			entryNumber--
		}

		slices.Reverse(memoryMessages)
		chatReq.Messages = append(chatReq.Messages, memoryMessages...)
	}

	content := strings.Builder{}

	content.WriteString(fmt.Sprintf(localize.Character["respond_as"], character.Name))

	if localize.Character["response_language"] != "" {
		content.WriteString(localize.Character["response_language"])
		content.WriteString(localize.Character["dont_say_response_language"])
	}

	if len(character.DontSay) > 0 || len(profile.DontSay) > 0 || len(profile.ReplaceWords["_"]) > 0 {
		dontSay := make([]string, 0, len(character.DontSay)+len(profile.DontSay)+len(profile.ReplaceWords["_"]))
		dontSay = append(dontSay, character.DontSay...)
		dontSay = append(dontSay, profile.DontSay...)
		dontSay = append(dontSay, profile.ReplaceWords["_"]...)

		if character.DontSayName {
			dontSay = append(dontSay, profile.Name)
		}

		content.WriteString(fmt.Sprintf(localize.Character["dont_say"], strings.Join(dontSay, ", ")))
	}

	if len(profile.TopicsDiscourage) > 0 {
		content.WriteString(fmt.Sprintf(localize.Character["topics_discourage"], strings.Join(profile.TopicsDiscourage, ", ")))
	}

	if !character.DontSayName && rand.Intn(5) == 0 {
		content.WriteString(fmt.Sprintf(localize.Character["refer_to_me_as"], profile.Name))
	}

	if profile.ResponseAge > 0 {
		content.WriteString(fmt.Sprintf(localize.Character["response_age"], profile.ResponseAge))
	}

	addQuestion := profile.Characters[character.Character].Mode == "conversation" &&
		profile.AddQuestionFrequency > 0 &&
		rand.Intn(100) <= profile.AddQuestionFrequency

	if addQuestion {
		content.WriteString(localize.Character["follow_up_question"])
		mode.MaxWords += 5
	}

	content.WriteString(fmt.Sprintf(localize.Character["max_words_1"], mode.MaxWords))
	content.WriteString(userPrompt)
	content.WriteString(fmt.Sprintf(localize.Character["max_words_2"], mode.MaxWords))

	if len(profile.Interests) > 0 && rand.Intn(10) == 0 {
		interest := profile.Interests[rand.Intn(len(profile.Interests))]
		content.WriteString(fmt.Sprintf(localize.Character["interests"], interest))
	}

	if len(profile.TopicsEncourage) > 0 && rand.Intn(5) == 0 {
		topic := profile.TopicsEncourage[rand.Intn(len(profile.TopicsEncourage))]
		content.WriteString(fmt.Sprintf(localize.Character["topics_encourage"], topic))
	}

	content.WriteString(localize.Character["dont_say_ai"])

	contentString := content.String()
	chatReq.Messages = append(chatReq.Messages, openai.ChatMessage{
		Role:    "user",
		Content: contentString,
	})

	// tokenize the request.
	tke, err := tiktoken.GetEncoding(characters.TokenEncodingModel)
	if err != nil {
		return nil, 0
	}
	numTokens := len(tke.Encode(systemPromptString+contentString, nil, nil))

	return &chatReq, numTokens
}

func conversationGPTPromptBuilderV2(profile *profiles.Document, character *characters.Character, mode *characters.Mode,
	session *characters.SessionDocument, userPrompt string, localize *configs.Localize, predefined bool) (*openai.ChatRequest, int) {
	systemPrompt := strings.Builder{}
	systemPrompt.WriteString(mode.CharacterPrompt)

	if len(character.TraitsPositive) > 0 {
		systemPrompt.WriteString(localize.Character["traits_positive"])
		systemPrompt.WriteString(strings.Join(character.TraitsPositive, ", "))
		systemPrompt.WriteString(".")
	}

	if len(character.TraitsNegative) > 0 {
		systemPrompt.WriteString(localize.Character["traits_negative"])
		systemPrompt.WriteString(strings.Join(character.TraitsNegative, ", "))
		systemPrompt.WriteString(".")
	}

	systemPrompt.WriteString(localize.Character["guardrail_prompt"])

	systemPromptString := systemPrompt.String()
	chatReq := openai.ChatRequest{
		Model:      character.Model,
		Messages:   []openai.ChatMessage{{Role: "system", Content: systemPromptString}},
		Creativity: mode.Creativity,
	}

	start := len(session.Entries) - mode.SessionEntries + session.StartEntry
	if start < 1 {
		start = 1
	}

	if !predefined {
		entryNumber := len(session.Entries) + session.StartEntry
		numEntries := 0
		memoryMessages := []openai.ChatMessage{}
		for numEntries <= mode.SessionEntries && entryNumber >= session.StartEntry {
			iStr := fmt.Sprintf("%06d", entryNumber)

			if session.Entries[iStr].EndSequence {
				break
			}

			if session.Entries[iStr].Moderation != nil && !session.Entries[iStr].Moderation.Triggered {
				memoryMessages = append(memoryMessages, openai.ChatMessage{Role: "assistant", Content: session.Entries[iStr].Assistant})
				memoryMessages = append(memoryMessages, openai.ChatMessage{Role: "user", Content: session.Entries[iStr].User})
				numEntries++
			}

			entryNumber--
		}

		slices.Reverse(memoryMessages)
		chatReq.Messages = append(chatReq.Messages, memoryMessages...)
	}

	content := strings.Builder{}

	content.WriteString(fmt.Sprintf(localize.Character["respond_as"], character.Name))

	if localize.Character["response_language"] != "" {
		content.WriteString(localize.Character["response_language"])
		content.WriteString(localize.Character["dont_say_response_language"])
	}

	if len(character.DontSay) > 0 || len(profile.DontSay) > 0 || len(profile.ReplaceWords["_"]) > 0 {
		dontSay := make([]string, 0, len(character.DontSay)+len(profile.DontSay)+len(profile.ReplaceWords["_"]))
		dontSay = append(dontSay, character.DontSay...)
		dontSay = append(dontSay, profile.DontSay...)
		dontSay = append(dontSay, profile.ReplaceWords["_"]...)

		if character.DontSayName {
			dontSay = append(dontSay, profile.Name)
		}

		content.WriteString(fmt.Sprintf(localize.Character["dont_say"], strings.Join(dontSay, ", ")))
	}

	if len(profile.TopicsDiscourage) > 0 {
		content.WriteString(fmt.Sprintf(localize.Character["topics_discourage"], strings.Join(profile.TopicsDiscourage, ", ")))
	}

	if !character.DontSayName && rand.Intn(5) == 0 {
		content.WriteString(fmt.Sprintf(localize.Character["refer_to_me_as"], profile.Name))
	}

	if profile.ResponseAge > 0 {
		content.WriteString(fmt.Sprintf(localize.Character["response_age"], profile.ResponseAge))
	}

	addQuestion := profile.Characters[character.Character].Mode == "conversation" &&
		profile.AddQuestionFrequency > 0 &&
		rand.Intn(100) <= profile.AddQuestionFrequency

	if addQuestion {
		content.WriteString(localize.Character["follow_up_question"])
		mode.MaxWords += 5
	}

	content.WriteString(fmt.Sprintf(localize.Character["max_words_1"], mode.MaxWords))
	content.WriteString(userPrompt)
	content.WriteString(fmt.Sprintf(localize.Character["max_words_2"], mode.MaxWords))

	if len(profile.Interests) > 0 && rand.Intn(10) == 0 {
		interest := profile.Interests[rand.Intn(len(profile.Interests))]
		content.WriteString(fmt.Sprintf(localize.Character["interests"], interest))
	}

	if len(profile.TopicsEncourage) > 0 && rand.Intn(5) == 0 {
		topic := profile.TopicsEncourage[rand.Intn(len(profile.TopicsEncourage))]
		content.WriteString(fmt.Sprintf(localize.Character["topics_encourage"], topic))
	}

	contentString := content.String()
	chatReq.Messages = append(chatReq.Messages, openai.ChatMessage{
		Role:    "user",
		Content: contentString,
	})

	// tokenize the request to determine 4k or 16k model.
	tke, err := tiktoken.GetEncoding(characters.TokenEncodingModel)
	if err != nil {
		return nil, 0
	}
	numTokens := len(tke.Encode(systemPromptString+contentString, nil, nil))

	return &chatReq, numTokens
}

func funGPTPromptBuilderV1(profile *profiles.Document, character *characters.Character, mode *characters.Mode,
	session *characters.SessionDocument, userPrompt string, localize *configs.Localize) (*openai.ChatRequest, int) {
	systemPrompt := strings.Builder{}
	systemPrompt.WriteString(mode.CharacterPrompt)

	if len(character.TraitsPositive) > 0 {
		systemPrompt.WriteString(localize.Character["traits_positive"])
		systemPrompt.WriteString(strings.Join(character.TraitsPositive, ", "))
		systemPrompt.WriteString(".")
	}

	if len(character.TraitsNegative) > 0 {
		systemPrompt.WriteString(localize.Character["traits_negative"])
		systemPrompt.WriteString(strings.Join(character.TraitsNegative, ", "))
		systemPrompt.WriteString(".")
	}

	systemPromptString := systemPrompt.String()
	chatReq := openai.ChatRequest{
		Model:      character.Model,
		Messages:   []openai.ChatMessage{{Role: "system", Content: systemPromptString}},
		Creativity: mode.Creativity,
	}

	entryNumber := len(session.Entries) + session.StartEntry
	numEntries := 0
	memoryMessages := []openai.ChatMessage{}
	for numEntries <= mode.SessionEntries && entryNumber >= session.StartEntry {
		iStr := fmt.Sprintf("%06d", entryNumber)

		if session.Entries[iStr].EndSequence {
			break
		}

		if session.Entries[iStr].Moderation != nil && !session.Entries[iStr].Moderation.Triggered {
			memoryMessages = append(memoryMessages, openai.ChatMessage{Role: "assistant", Content: session.Entries[iStr].Assistant})
			memoryMessages = append(memoryMessages, openai.ChatMessage{Role: "user", Content: session.Entries[iStr].User})
			numEntries++
		}

		entryNumber--
	}

	slices.Reverse(memoryMessages)
	chatReq.Messages = append(chatReq.Messages, memoryMessages...)

	content := strings.Builder{}

	content.WriteString(fmt.Sprintf(localize.Character["respond_as"], character.Name))

	if localize.Character["response_language"] != "" {
		content.WriteString(localize.Character["response_language"])
		content.WriteString(localize.Character["dont_say_response_language"])
	}

	if len(character.DontSay) > 0 || len(profile.DontSay) > 0 || len(profile.ReplaceWords["_"]) > 0 {
		dontSay := make([]string, 0, len(character.DontSay)+len(profile.DontSay)+len(profile.ReplaceWords["_"]))
		dontSay = append(dontSay, character.DontSay...)
		dontSay = append(dontSay, profile.DontSay...)
		dontSay = append(dontSay, profile.ReplaceWords["_"]...)

		if character.DontSayName {
			dontSay = append(dontSay, profile.Name)
		}

		content.WriteString(fmt.Sprintf(localize.Character["dont_say"], strings.Join(dontSay, ", ")))
	}

	if profile.ResponseAge > 0 {
		content.WriteString(fmt.Sprintf(localize.Character["response_age"], profile.ResponseAge))
	}

	content.WriteString(fmt.Sprintf(localize.Character["max_words_1"], mode.MaxWords))
	content.WriteString(userPrompt)
	content.WriteString(fmt.Sprintf(localize.Character["max_words_2"], mode.MaxWords))

	content.WriteString(localize.Character["dont_say_ai"])

	contentString := content.String()
	chatReq.Messages = append(chatReq.Messages, openai.ChatMessage{
		Role:    "user",
		Content: contentString,
	})

	// tokenize the request.
	tke, err := tiktoken.GetEncoding(characters.TokenEncodingModel)
	if err != nil {
		return nil, 0
	}
	numTokens := len(tke.Encode(systemPromptString+contentString, nil, nil))

	return &chatReq, numTokens
}

func funGPTPromptBuilderV2(profile *profiles.Document, character *characters.Character, mode *characters.Mode,
	session *characters.SessionDocument, userPrompt string, localize *configs.Localize) (*openai.ChatRequest, int) {
	systemPrompt := strings.Builder{}
	systemPrompt.WriteString(mode.CharacterPrompt)

	if len(character.TraitsPositive) > 0 {
		systemPrompt.WriteString(localize.Character["traits_positive"])
		systemPrompt.WriteString(strings.Join(character.TraitsPositive, ", "))
		systemPrompt.WriteString(".")
	}

	if len(character.TraitsNegative) > 0 {
		systemPrompt.WriteString(localize.Character["traits_negative"])
		systemPrompt.WriteString(strings.Join(character.TraitsNegative, ", "))
		systemPrompt.WriteString(".")
	}

	systemPrompt.WriteString(localize.Character["guardrail_prompt"])

	systemPromptString := systemPrompt.String()
	chatReq := openai.ChatRequest{
		Model:      character.Model,
		Messages:   []openai.ChatMessage{{Role: "system", Content: systemPromptString}},
		Creativity: mode.Creativity,
	}

	entryNumber := len(session.Entries) + session.StartEntry
	numEntries := 0
	memoryMessages := []openai.ChatMessage{}
	for numEntries <= mode.SessionEntries && entryNumber >= session.StartEntry {
		iStr := fmt.Sprintf("%06d", entryNumber)

		if session.Entries[iStr].EndSequence {
			break
		}

		if session.Entries[iStr].Moderation != nil && !session.Entries[iStr].Moderation.Triggered {
			memoryMessages = append(memoryMessages, openai.ChatMessage{Role: "assistant", Content: session.Entries[iStr].Assistant})
			memoryMessages = append(memoryMessages, openai.ChatMessage{Role: "user", Content: session.Entries[iStr].User})
			numEntries++
		}

		entryNumber--
	}

	slices.Reverse(memoryMessages)
	chatReq.Messages = append(chatReq.Messages, memoryMessages...)

	content := strings.Builder{}

	content.WriteString(fmt.Sprintf(localize.Character["respond_as"], character.Name))

	if localize.Character["response_language"] != "" {
		content.WriteString(localize.Character["response_language"])
		content.WriteString(localize.Character["dont_say_response_language"])
	}

	if len(character.DontSay) > 0 || len(profile.DontSay) > 0 || len(profile.ReplaceWords["_"]) > 0 {
		dontSay := make([]string, 0, len(character.DontSay)+len(profile.DontSay)+len(profile.ReplaceWords["_"]))
		dontSay = append(dontSay, character.DontSay...)
		dontSay = append(dontSay, profile.DontSay...)
		dontSay = append(dontSay, profile.ReplaceWords["_"]...)

		if character.DontSayName {
			dontSay = append(dontSay, profile.Name)
		}

		content.WriteString(fmt.Sprintf(localize.Character["dont_say"], strings.Join(dontSay, ", ")))
	}

	if profile.ResponseAge > 0 {
		content.WriteString(fmt.Sprintf(localize.Character["response_age"], profile.ResponseAge))
	}

	content.WriteString(fmt.Sprintf(localize.Character["max_words_1"], mode.MaxWords))
	content.WriteString(userPrompt)
	content.WriteString(fmt.Sprintf(localize.Character["max_words_2"], mode.MaxWords))

	content.WriteString(localize.Character["dont_say_ai"])

	contentString := content.String()
	chatReq.Messages = append(chatReq.Messages, openai.ChatMessage{
		Role:    "user",
		Content: contentString,
	})

	// tokenize the request.
	tke, err := tiktoken.GetEncoding(characters.TokenEncodingModel)
	if err != nil {
		return nil, 0
	}
	numTokens := len(tke.Encode(systemPromptString+contentString, nil, nil))

	return &chatReq, numTokens
}
