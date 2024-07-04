package llm

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// ConvertCSVToJSONL converts a CSV file to JSONL format.
func ConvertCSVToJSONL(ctx context.Context, csvFilename string) error {
	_ = ctx
	logCtx := slog.With("filename", csvFilename)
	logCtx.Info("Converting to jsonl...")

	// Create input file
	inF, err := os.Open(csvFilename)
	if err != nil {
		logCtx.Error("unable to open csv file", "error", err)
		return err
	}
	defer inF.Close()

	// Create output file
	outF, err := os.Create(strings.Replace(csvFilename, ".csv", ".jsonl", 1))
	if err != nil {
		logCtx.Error("unable to create jsonl file", "error", err)
		return err
	}
	defer outF.Close()

	r := csv.NewReader(inF)
	rows, err := r.ReadAll()
	if err != nil {
		logCtx.Error("unable to read file", "error", err)
		return errors.New("invalid file")
	}
	if len(rows) < 3 {
		logCtx.Error("invalid number of rows")
		return errors.New("invalid number of rows")
	}

	headers := map[string]int{}

	for i, c := range rows[0] {
		headers[c] = i
	}

	if headers["prompt"] == 0 || headers["completion"] == 0 {
		logCtx.Error("invalid columns")
		return errors.New("invalid columns")
	}

	replacer := strings.NewReplacer(
		"“", "'",
		"”", "'",
		"‘", "'",
		"’", "'",
	)

	for i := 1; i < len(rows); i++ {
		prompt := replacer.Replace(strings.TrimSpace(rows[i][headers["prompt"]])) + "\n\n###\n\n"
		completion := replacer.Replace(strings.TrimSpace(rows[i][headers["compeltion"]])) + "\n"

		_, err := outF.WriteString(fmt.Sprintf("{\"prompt\": %q, \"completion\": %q}\n", prompt, completion))
		if err != nil {
			logCtx.Error("unable to write to file", "error", err)
			return err
		}
	}

	return nil
}
