package openai

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"disruptive/lib/common"
)

// ChatMessage contains a single chat message.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatRequest is a the request API structure.
type ChatRequest struct {
	Model      string        `json:"model"`
	Messages   []ChatMessage `json:"messages"`
	Creativity int           `json:"creativity"`
	MaxTokens  int           `json:"max_tokens"`
}

type chatResponse struct {
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type chatChunkResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// ChatResponse is the response API structure.
type ChatResponse struct {
	Text          string `json:"text"`
	FinishReason  string `json:"finish_reason"`
	Usage         int    `json:"usage"`
	UsagePrompt   int    `json:"usage_prompt"`
	UsageResponse int    `json:"usage_response"`
}

func renderFromChatReq(req ChatRequest) chatRequest {
	if req.Creativity < 0 {
		req.Creativity = 0
	}

	if req.Creativity > 100 {
		req.Creativity = 100
	}

	return chatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: float64(req.Creativity) * 2 / 100,
		MaxTokens:   req.MaxTokens,
	}
}

func renderToChatRes(_res *chatResponse) ChatResponse {
	text := ""
	reason := ""
	if len(_res.Choices) > 0 {
		text = _res.Choices[0].Message.Content
		reason = _res.Choices[0].FinishReason
	}

	return ChatResponse{
		Text:          text,
		FinishReason:  reason,
		UsagePrompt:   _res.Usage.PromptTokens,
		UsageResponse: _res.Usage.CompletionTokens,
	}
}

// PostChat sends chat prompts to OpenAI.
func PostChat(ctx context.Context, logCtx *slog.Logger, req ChatRequest) (ChatResponse, error) {
	fid := slog.String("fid", "openai.PostChat")

	_req := renderFromChatReq(req)

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(_req).
		SetResult(&chatResponse{}).
		Post(chatEndpoint)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logCtx.Error("timeout", "error", err)
			return ChatResponse{}, err
		}
		logCtx.Error("openai chat endpoint failed", fid, "error", err)
		return ChatResponse{}, err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("openai unauthorized", fid, "status", res.Status())
		return ChatResponse{}, common.ErrUnauthorized
	}

	if res.StatusCode() == http.StatusBadGateway {
		logCtx.Error("openai chat endpoint bad gateway", fid, "status", res.Status())
		return ChatResponse{}, common.ErrConnection{Msg: "openai: " + res.Status()}
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("openai chat endpoint failed", fid, "status", res.Status())
		return ChatResponse{}, errors.New("openai: " + res.Status())
	}

	return renderToChatRes(res.Result().(*chatResponse)), nil
}

// PostChatStream sends chat prompts to OpenAI.
func PostChatStream(ctx context.Context, logCtx *slog.Logger, req ChatRequest) (io.Reader, error) {
	fid := slog.String("fid", "openai.PostChatStream")

	_req := renderFromChatReq(req)
	_req.Stream = true

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(_req).
		SetDoNotParseResponse(true).
		Post(chatEndpoint)

	if err != nil {
		logCtx.Error("openai chat endpoint failed", fid, "error", err)
		return nil, err
	}

	r, w := io.Pipe()

	go func() {
		defer res.RawBody().Close()
		defer w.Close()

		scanner := bufio.NewScanner(res.RawBody())
		for scanner.Scan() {
			chunk := scanner.Bytes()
			chunkLen := len(chunk)
			if chunkLen < 6 {
				continue
			}

			chunk = chunk[6:]

			if chunk[0] == '[' { // [DONE]
				continue
			}

			res := chatChunkResponse{}

			if err := json.Unmarshal(chunk, &res); err != nil {
				logCtx.Error("unable to read chunk data", fid, "error", err)
				return
			}

			if len(res.Choices) < 1 {
				logCtx.Warn("no response choices")
				continue
			}

			if _, err := w.Write([]byte(res.Choices[0].Delta.Content)); err != nil {
				logCtx.Error("unable to write to chunk pipe", fid, "error", err)
				return
			}
		}

		if err := scanner.Err(); err != nil {
			logCtx.Error("unable to scan response", fid, "error", err)
			return
		}
	}()

	return r, nil
}
