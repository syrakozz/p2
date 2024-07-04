package configs

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"

	"cloud.google.com/go/firestore"

	"disruptive/config"
	"disruptive/lib/configs"
)

// PutPersonalities sets the personalities config file.
// See: data/system-prompts.csv
func PutPersonalities(ctx context.Context, filename string) error {
	logCtx := slog.With("filename", filename)

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return err
	}
	defer client.Close()

	document := configs.Document{}

	f, err := os.Open(filename)
	if err != nil {
		logCtx.Error("unable open file", "error", err)
		return err
	}
	r := csv.NewReader(f)

	h, err := r.Read()
	if err != nil {
		logCtx.Error("unable read file", "error", err)
		return err
	}

	if len(h) < 2 || h[0] != "name" || h[1] != "value" {
		logCtx.Error("invalid file format", "error", err)
		return err
	}

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			logCtx.Error("unable read file", "error", err)
			return err
		}

		if len(row) != 2 {
			logCtx.Error("invalid file format")
			return err
		}

		document[row[0]] = row[1]

	}

	fmt.Println(document)

	configs.Put(ctx, logCtx, "personalities", document)

	return nil
}
