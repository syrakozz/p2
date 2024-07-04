package llm

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"disruptive/lib/common"
	"disruptive/lib/openai"
	"disruptive/lib/pinecone"
)

// UpsertVectors creates or updates Pinecone vectors
func UpsertVectors(ctx context.Context, namespace, filePath string, continueFlag bool) error {
	logCtx := slog.With()

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if len(rows) < 2 {
		return nil
	}

	seen := map[string]struct{}{}

	if continueFlag {
		func() {
			seenF, err := os.Open("seen.csv")
			if err != nil {
				return
			}
			defer seenF.Close()

			rows, err := csv.NewReader(seenF).ReadAll()
			if err != nil {
				logCtx.Error("unable to read seen.csv", "error", err)
				return
			}

			for _, r := range rows {
				seen[r[0]] = struct{}{}
			}
		}()
	}

	var fSeen *os.File

	if continueFlag {
		fSeen, err = os.OpenFile("seen.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	} else {
		fSeen, err = os.Create("seen.csv")
	}
	if err != nil {
		logCtx.Error("unable to create seen file", "error", err)
		return err
	}
	defer fSeen.Close()

	csvSeen := csv.NewWriter(fSeen)
	defer csvSeen.Flush()

	cols := map[string]int{}
	for i, c := range rows[0] {
		cols[c] = i
	}

	// Replace smart quotes
	replacer := strings.NewReplacer(
		"“", "'",
		"”", "'",
		"‘", "'",
		"’", "'",
	)

	for i, row := range rows[1:] {
		select {
		case <-ctx.Done():
			logCtx.Info("canceled")
			return context.Canceled
		default:
		}

		if row[cols["pinecone"]] != "" {
			continue
		}

		uuid := replacer.Replace(strings.TrimSpace(row[cols["uuid"]]))
		question := replacer.Replace(strings.TrimSpace(row[cols["question"]]))
		answer := replacer.Replace(strings.TrimSpace(row[cols["answer"]]))

		if uuid == "" || question == "" || answer == "" {
			continue
		}

		if continueFlag {
			if _, ok := seen[uuid]; ok {
				continue
			}
		}

		embeddingsReq := openai.EmbeddingsRequest{
			Model: "text-embedding-ada-002",
			Input: []string{fmt.Sprintf("%s %s", question, answer)},
		}

		embeddingsRes, err := openai.PostEmbeddings(ctx, logCtx, embeddingsReq)
		if err != nil {
			logCtx.Error("unable to create embedding", "error", err)
			return err
		}

		if len(embeddingsRes.Data) != 1 {
			logCtx.Error("invalid embedding data", "error", err)
			return errors.New("invalid embeddings data")
		}

		upsertReq := pinecone.UpsertRequest{
			Namespace: namespace,
			Vectors: []pinecone.Vector{
				{
					ID:     uuid,
					Values: embeddingsRes.Data[0].Embedding,
					Metadata: pinecone.Metadata{
						"question": question,
						"answer":   answer,
					},
				},
			},
		}

		ids, err := pinecone.Upsert(ctx, logCtx, upsertReq)
		if err != nil {
			logCtx.Error("Invalid upsert", "uuid", uuid, "row", i+1)
			return err
		}

		if len(ids) < 1 {
			logCtx.Error("Invalid upsert", "uuid", uuid, "row", i+1)
			return errors.New("invalid upsert")
		}

		csvSeen.Write([]string{uuid})
		csvSeen.Flush()

		logCtx.Info("Processed", "id", ids[0], "row", i+1)
	}

	logCtx.Info("upsert complete")
	return nil
}

// DeleteVectors delete Pinecone vectors by IDs or all within a single namespace.
func DeleteVectors(ctx context.Context, namespace string, ids []string, deleteAll bool) error {
	logCtx := slog.With("num_ids", len(ids), "deleteAll", deleteAll)

	request := pinecone.DeleteRequest{
		Namespace: namespace,
		IDs:       ids,
		DeleteAll: deleteAll,
	}

	if err := pinecone.Delete(ctx, logCtx, request); err != nil {
		logCtx.Error("unable to delete", "error", err)
		return err
	}

	logCtx.Info("deleted")
	return nil
}

// StatsVectors retrieves the Pinecone namespace status.
func StatsVectors(ctx context.Context) error {
	logCtx := slog.With()

	res, err := pinecone.Stats(ctx, logCtx, pinecone.Metadata{})
	if err != nil {
		logCtx.Error("unable to retrieve stats", "error", err)
		return err
	}

	common.P(res)
	return nil
}

// QueryVectors displays the closest topK vectors in Pinecone from the query.
func QueryVectors(ctx context.Context, namespace, query string, topK int) error {
	logCtx := slog.With("namespace", namespace, "query", query, "topK", topK)

	embeddingsReq := openai.EmbeddingsRequest{
		Model: "text-embedding-ada-002",
		Input: []string{query},
	}

	embeddingsRes, err := openai.PostEmbeddings(ctx, logCtx, embeddingsReq)
	if err != nil {
		logCtx.Error("unable to create embedding", "error", err)
		return err
	}

	if len(embeddingsRes.Data) != 1 {
		logCtx.Error("invalid embedding data", "error", err)
		return errors.New("invalid embeddings data")
	}

	queryReq := pinecone.QueryRequest{
		Namespace:       namespace,
		TopK:            topK,
		Vector:          embeddingsRes.Data[0].Embedding,
		IncludeMetadata: true,
	}

	queryRes, err := pinecone.Query(ctx, logCtx, queryReq)
	if err != nil {
		logCtx.Error("unable to query", "error", err)
		return err
	}

	common.P(queryRes)
	return nil
}

// FetchVector fetches a vector by ID.
func FetchVector(ctx context.Context, namespace, id string) error {
	logCtx := slog.With("namespace", namespace, "id", id)

	req := pinecone.FetchRequest{
		Namespace: namespace,
		IDs:       []string{id},
	}

	res, err := pinecone.Fetch(ctx, logCtx, req)
	if err != nil {
		logCtx.Error("unable to fetch document", "error", err)
		return err
	}

	common.P(res)
	return nil
}
