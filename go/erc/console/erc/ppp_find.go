package erc

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"disruptive/config"
)

type includeFunc func(h map[string]int, row []string) (bool, error)

// Find records.
func Find(ctx context.Context, maxNum int, size string) error {
	if maxNum < 1 {
		return nil
	}

	id := time.Now().Format("060102T150405")

	if err := os.MkdirAll(filepath.Join(config.VARS.ERCDataRoot, "process"), os.ModePerm); err != nil {
		return err
	}

	switch {
	case size != "":
		return findSize(ctx, id, maxNum, size)
	default:
		return errors.New("invaid find criteria")
	}

}
