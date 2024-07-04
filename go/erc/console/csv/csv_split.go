// Package csv contains csv file operations
package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SplitCSV splits a CSV file into sub-files preversing the header.
func SplitCSV(ctx context.Context, num int, file string) error {
	fin, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fin.Close()

	ext := filepath.Ext(file)
	withoutExt := strings.TrimSuffix(file, ext)

	files := make([]*os.File, num)
	csvs := make([]*csv.Writer, num)

	for i := 0; i < num; i++ {
		f, err := os.Create(fmt.Sprintf("%s-%02d%s", withoutExt, i+1, ext))
		if err != nil {
			return err
		}
		defer f.Close()

		files[i] = f

		csvs[i] = csv.NewWriter(f)
		defer csvs[i].Flush()
	}

	r := csv.NewReader(fin)

	// read header

	header, err := r.Read()
	if err != nil {
		return err
	}
	header = append(header, "Sub-Account")

	// write header

	for i := 0; i < num; i++ {
		if err := csvs[i].Write(header); err != nil {
			return err
		}
	}

	i := 0
	for {
		row, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		row = append(row, fmt.Sprintf("%02d", i%num+1))
		if err := csvs[i%num].Write(row); err != nil {
			return err
		}

		i++
	}

	return nil
}
