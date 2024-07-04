package mbox

import (
	"context"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"disruptive/lib/openai"
	"disruptive/lib/pinecone"
)

// PineconeFunc prints each mbox part to the console.
// Return true to continue processing the next email.
func PineconeFunc(ctx context.Context, logCtx *slog.Logger, namespace string, mediaType string, header mail.Header, body string) bool {
	logCtx = logCtx.With("len", len(body))

	// Not used
	_ = mediaType

	var fromStr string

	// only process entries with a messageID.
	messageID := header.Get("Message-ID")
	if messageID == "" {
		logCtx.Warn("missing Message-ID")
		return true
	}

	id := cleanID.Replace(messageID)

	from, _ := header.AddressList("From")
	if len(from) > 0 {
		fromStr = from[0].String()
	}

	to, _ := header.AddressList("To")
	toStrs := []string{}
	for _, t := range to {
		toStrs = append(toStrs, t.String())
	}

	cc, _ := header.AddressList("Cc")
	ccStrs := []string{}
	for _, c := range cc {
		ccStrs = append(ccStrs, c.String())
	}

	bcc, _ := header.AddressList("Bcc")
	bccStrs := []string{}
	for _, b := range bcc {
		bccStrs = append(bccStrs, b.String())
	}

	date, _ := header.Date()
	datetime := date.UTC().Format(time.DateTime)

	subject := decodeHeader(header.Get("Subject"))
	threadTopic := header.Get("Thread-Topic")
	if threadTopic != "" {
		subject = threadTopic
	}

	var body10k string
	if len(body) > 10000 {
		body10k = body[:10000]
	}

	logCtx = logCtx.With("id", id, "subject", subject)

	embeddingReq := openai.EmbeddingsRequest{
		Model: "text-embedding-ada-002",
		Input: []string{fmt.Sprintf("%s\n%s", subject, body10k)},
	}

	embeddingRes, err := openai.PostEmbeddings(ctx, logCtx, embeddingReq)
	if err != nil {
		logCtx.Error("unable to post embeddings", "error", err)
		return false
	}

	if len(embeddingRes.Data) != 1 {
		logCtx.Error("inconsistent embedding data length", "len", len(embeddingRes.Data))
		return false
	}

	upsertReq := pinecone.UpsertRequest{
		Namespace: namespace,
		Vectors: []pinecone.Vector{
			{
				ID: id,
				Metadata: pinecone.Metadata{
					"from":     fromStr,
					"to":       strings.Join(toStrs, ", "),
					"cc":       strings.Join(ccStrs, ", "),
					"bcc":      strings.Join(bccStrs, ", "),
					"datetime": datetime,
					"subject":  subject,
					"body":     body,
				},
				Values: embeddingRes.Data[0].Embedding,
			},
		},
	}

	pinecone.Upsert(ctx, logCtx, upsertReq)

	logCtx.Info("upsert")

	return true

}
