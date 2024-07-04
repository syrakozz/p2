package erc

import (
	"context"
	"disruptive/config"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type (
	pppHeader           []string
	pppHeaderIndexes    map[string]int
	pppRowsByLoanNumber map[string][]string
)

func mergePpp(ctx context.Context, basename string) (pppHeader, pppHeaderIndexes, pppRowsByLoanNumber, error) {
	logCtx := log.WithField("basename", basename)

	path := filepath.Join(config.VARS.ERCDataRoot, "process", basename+"-ppp.csv")

	f, err := os.Open(path)
	if err != nil {
		logCtx.WithError(err).Error("unable to open header")
		return nil, nil, nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)

	// read header

	h, err := r.Read()
	if err != nil {
		logCtx.WithError(err).Error("unable to read line")
		return nil, nil, nil, err
	}

	headers := make(map[string]int, len(h))
	for i, h := range h {
		headers[h] = i
	}

	// read rows

	rows := make(pppRowsByLoanNumber, 50000)

	for {
		select {
		case <-ctx.Done():
			log.Warn("canceled")
			return nil, nil, nil, context.Canceled
		default:
		}

		row, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			logCtx.WithError(err).Error("unable to read file")
			return nil, nil, nil, err
		}

		rows[row[headers["LoanNumber"]]] = row
	}

	return h, headers, rows, nil
}
