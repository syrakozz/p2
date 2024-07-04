// Package moderate retrieves moderation from all moderation services.
package moderate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/openai"
)

const (
	promptTemplate = "Please analyze the following text for its appropriateness for children. " +
		"Consider the language used, the complexity of the vocabulary and sentences, the themes and topics discussed, " +
		"and any sensitive or mature content. Provide a detailed assessment indicating whether the text is suitable for children, " +
		"and if so, suggest an appropriate age. Additionally, highlight any specific parts of the text that might be concerning " +
		"and explain why they may not be suitable for younger audiences. Statements that are real facts or general questions should be allowed for young ages. " +
		"Brand names, commercial platforms, technology, common slang, and informal language should be allowed for young ages. " +
		"Allow the use of the words 'XL' and 'Excel'. " +
		"Only generate the output result in JSON without any explanations using the following fields: " +
		"'assessment', 'assessment translation' in the language %s, 'assessment age' (integer), 'movie rating', 'TV rating', 'Entertainment Software Rating Board (ESRB)', 'Pan European Game Information (PEGI) rating', " +
		"'sentiment', 'emotion', 'intent classification', 'toxicity detection', named entities of people and places, " +
		"ISO-639-1 language code, and 'subject category' of the following INPUT. Write the response in RFC8259 compliant JSON using these examples:" +
		`
	
EXAMPLE 1:

Why is the sky blue?

{
    "assessment": "General science question.",
    "assessment_translation": "Pregunta de ciencia general.",
    "assessment_age": 0,
	"topic": "science",
    "classifications": ["positive", "educational", "ethical"],
    "movie_rating": "G",
    "tv_rating": "TV-Y",
    "esrb_rating": "E",
    "pegi_rating": 3,
	"sentiment": "positive",
	"emotion": "neutral",
	"intent": "question",
	"toxic": false,
	"named_entities": [],
	"subject_category": "science",
	"language": "en"
}

EXAMPLE 2:

Baseball is a very fun sport, but players sometimes get hurt.

{
	"assessment": "Suitable for children. The text expresses a subjective opinion about a sport and mentions players getting hurt, which may need explanation for very young children.",
    "assessment_translation": "Convient aux enfants. Le texte exprime une opinion subjective sur un sport et mentionne des blessures chez des joueurs, ce qui peut nécessiter une explication pour les très jeunes enfants.",
	"assessment_age": 3,
    "topic": "sports",
    "classifications": ["educational", "sport"],
    "movie_rating": "G",
    "tv_rating": "TV-Y",
    "esrb_rating": "E",
    "pegi_rating": 3,
	"sentiment": "neutral",
	"emotion": "negative",
	"intent": "statement",
	"toxic": false,
	"named_entities": ["Baseball"],
	"subject_category": "Lifestyle",
	"language": "en"
}

EXAMPLE 3:

I love Genghis Khan.

{
	"assessment": "Not suitable for children. The text mentions a historical figure associated with extreme violence and genocide, which is inappropriate and potentially disturbing for young readers.",
    "assessment_translation": "Nicht geeignet für Kinder. Der Text erwähnt eine historische Persönlichkeit, die mit extremer Gewalt und Völkermord in Verbindung gebracht wird, was für junge Leser unangemessen und möglicherweise verstörend ist.",
	"assessment_age": 13,
    "topic": "history",
    "classifications": [],
    "movie_rating": "PG-13",
    "tv_rating": "TV-14",
    "esrb_rating": "T",
    "pegi_rating": 12,
	"sentiment": "happy",
	"emotion": "positive",
	"intent": "statement",
	"toxic": true,
	"named_entities": ["Genghis Khan"],
	"subject_category": "history",
	"language": "en"
}

INPUT:
%s

JSON OUTPUT:
`
)

// Response contains moderation information from all services.
type Response struct {
	sync.Mutex
	Categories map[string]bool `firestore:"categories" json:"categories"`
	Analysis   struct {
		AssessmentAge         int      `firestore:"assessment_age,omitempty" json:"assessment_age,omitempty"`
		Assessment            string   `firestore:"assessment,omitempty" json:"assessment,omitempty"`
		AssessmentTranslation string   `firestore:"assessment_translation,omitempty" json:"assessment_translation,omitempty"`
		Classifications       []string `firestore:"classifications,omitempty" json:"classifications,omitempty"`
		Emotion               string   `firestore:"emotion,omitempty" json:"emotion,omitempty"`
		Entities              []string `firestore:"entities,omitempty" json:"named_entities,omitempty"`
		ESRBRating            string   `firestore:"esrb_rating,omitempty" json:"esrb_rating,omitempty"`
		Intent                string   `firestore:"intent,omitempty" json:"intent,omitempty"`
		Language              string   `firestore:"language,omitempty" json:"language,omitempty"`
		MovieRating           string   `firestore:"movie_rating,omitempty" json:"movie_rating,omitempty"`
		NotAgeAppropriate     bool     `firestore:"not_age_appropriate,omitempty" json:"not_age_appropriate,omitempty"`
		PEGIRating            int      `firestore:"pegi_rating,omitempty" json:"pegi_rating,omitempty"`
		Sentiment             string   `firestore:"sentiment,omitempty" json:"sentiment,omitempty"`
		SubjectCategory       string   `firestore:"subject_category,omitempty" json:"subject_category,omitempty"`
		Topic                 string   `firestore:"topic,omitempty" json:"topic,omitempty"`
		Toxic                 bool     `firestore:"toxic,omitempty" json:"toxic,omitempty"`
		TVRating              string   `firestore:"tv_rating,omitempty" json:"tv_rating,omitempty"`
	} `firestore:"analysis" json:"analysis"`
	TokensPrompt   int  `firestore:"tokens_prompt" json:"tokens_prompt,omitempty"`
	TokensResponse int  `firestore:"tokens_response" json:"tokens_response,omitempty"`
	Triggered      bool `firestore:"triggered" json:"triggered,omitempty"`
}

// Get return a moderation response for some input text.
func Get(ctx context.Context, logCtx *slog.Logger, text, language string) *Response {
	fid := slog.String("fid", "vox.moderate.Get")

	res := &Response{}

	g := errgroup.Group{}

	g.Go(func() error {
		return getCategories(ctx, logCtx, text, res)
	})

	g.Go(func() error {
		return getAnalysis(ctx, logCtx, text, language, res)
	})

	if err := g.Wait(); err != nil {
		logCtx.Warn("unable to process all moderations", fid, "error", err)
	}

	return res
}

func getCategories(ctx context.Context, logCtx *slog.Logger, text string, response *Response) error {
	fid := slog.String("fid", "vox.moderate.getCategories")

	res, err := openai.PostModeration(ctx, logCtx, text)
	if err != nil {
		logCtx.Warn("unable to get moderation", fid, "error", err)
		return err
	}

	if len(res.Results) != 1 {
		return common.ErrConsistency
	}

	response.Lock()
	response.Categories = res.Results[0].Categories
	response.Unlock()

	return nil
}

func getAnalysis(ctx context.Context, logCtx *slog.Logger, text, language string, response *Response) error {
	fid := slog.String("fid", "vox.moderate.getAnalysis")

	chatReq := openai.ChatRequest{
		Model:     "gpt-3.5-turbo",
		Messages:  []openai.ChatMessage{{Role: "user", Content: fmt.Sprintf(promptTemplate, language, text)}},
		MaxTokens: 250,
	}

	chatRes, err := openai.PostChat(ctx, logCtx, chatReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logCtx.Error("timeout", fid, "error", err)
			return err
		}
		logCtx.Error("unable to get chat response", fid, "error", err)
		return err
	}

	promptCost := float64(chatRes.UsagePrompt) * config.VARS.GPT35TurboPromptCost / 1000.0
	responseCost := float64(chatRes.UsageResponse) * config.VARS.GPT35TurboResponseCost / 1000.0

	logCtx.Info(
		"cost",
		"feature", "text-analysis",
		"model", "gpt-3.5-turbo",
		"tokens_prompt", chatRes.UsagePrompt,
		"tokens_response", chatRes.UsageResponse,
		"tokens_total", chatRes.UsagePrompt+chatRes.UsageResponse,
		"cost_prompt", fmt.Sprintf("%.7f", promptCost),
		"cost_response", fmt.Sprintf("%.7f", responseCost),
		"cost_total", fmt.Sprintf("%.7f", promptCost+responseCost),
	)

	response.Lock()
	json.Unmarshal([]byte(chatRes.Text), &response.Analysis)
	response.TokensPrompt = chatRes.UsagePrompt
	response.TokensResponse = chatRes.UsageResponse

	if strings.Contains(language, "en-") {
		response.Analysis.AssessmentTranslation = response.Analysis.Assessment
	}

	response.Unlock()

	return nil
}
