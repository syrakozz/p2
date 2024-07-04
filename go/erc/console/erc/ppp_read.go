package erc

import (
	"context"
	"encoding/csv"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"disruptive/config"
)

func readPPPFiles(ctx context.Context, logCtx *log.Entry, id string, maxNum int, include includeFunc) error {
	// read already processed LoanNumber rows

	seen, err := getSeenLoanNumbers(ctx, logCtx)
	if err != nil {
		return err
	}

	// read ppp files

	path := filepath.Join(config.VARS.ERCDataRoot, "ppp")

	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	rand.Shuffle(len(files), func(i, j int) { files[i], files[j] = files[j], files[i] })

	// prepare RocketReach files

	rr, err := newRocketReach(ctx, logCtx, id)
	if err != nil {
		return err
	}
	defer rr.close()

	num := 0

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		num, err = readPPPFile(ctx, id, filepath.Join(path, f.Name()), num, maxNum, seen, rr, include)
		if err != nil {
			return err
		}

		if num >= maxNum {
			return nil
		}
	}

	return nil
}

func readPPPFile(ctx context.Context, id, file string, num, maxNum int, seen map[string]struct{}, rr *rocketReach, include includeFunc) (int, error) {
	logCtx := log.WithField("file", file)
	logCtx.Info("Processing")

	f, err := os.Open(file)
	if err != nil {
		log.WithError(err).Error("unable to read file")
		return 0, err
	}
	defer f.Close()

	r := csv.NewReader(f)

	// read header

	row, err := r.Read()
	if err != nil {
		logCtx.WithError(err).Error("unable to read header")
		return 0, err
	}

	headers := make(map[string]int, len(row))
	for i, h := range row {
		headers[h] = i
	}

	// read rows

	cleanString := strings.NewReplacer(
		"&amp;amp;", "&",
		"&amp;apos;", "'",
		"&amp;", "&",
		"  ", " ",
	)

	for {
		select {
		case <-ctx.Done():
			log.Warn("canceled")
			return 0, context.Canceled
		default:
		}

		row, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			logCtx.WithError(err).Error("unable to read file")
			return 0, err
		}

		// check if this row has already been processed

		if _, ok := seen[row[headers["LoanNumber"]]]; ok {
			continue
		}

		// clean row values

		for i := 0; i < len(row); i++ {
			s := strings.TrimSpace(row[i])
			s = strings.Join(strings.Fields(s), " ")
			row[i] = cleanString.Replace(s)
		}

		ok, err := include(headers, row)
		if err != nil {
			return 0, err
		}

		if !ok {
			continue
		}

		if err := rr.write(num, row); err != nil {
			return 0, err
		}

		num++

		if num >= maxNum {
			return num, nil
		}
	}

	return num, nil
}

func getSeenLoanNumbers(ctx context.Context, logCtx *log.Entry) (map[string]struct{}, error) {
	seen := make(map[string]struct{}, 12_000_000)

	path := filepath.Join(config.VARS.ERCDataRoot, "process")

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), "-rocketreach-in.csv") {
			continue
		}

		f2, err := os.Open(filepath.Join(path, f.Name()))
		if err != nil {
			return nil, err
		}

		r := csv.NewReader(f2)
		// throw away header
		if _, err := r.Read(); err != nil {
			f2.Close()
			return nil, err
		}

		for {
			select {
			case <-ctx.Done():
				log.Warn("canceled")
				f2.Close()
				return nil, context.Canceled
			default:
			}

			row, err := r.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				logCtx.WithError(err).Error("unable to read file")

				f2.Close()
				return nil, err
			}

			seen[row[0]] = struct{}{}
		}
		f2.Close()

	}

	return seen, nil
}
