package erc

import (
	"context"
	"errors"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func findSize(ctx context.Context, id string, maxNum int, size string) error {
	logCtx := log.WithFields(log.Fields{"num": maxNum, "id": id})

	parts := strings.Split(size, "-")
	if len(parts) != 2 {
		err := errors.New("invalid size range")
		logCtx.WithError(err).WithField("size", size).Error("Find by size")
		return err
	}

	sizeMin, err := strconv.Atoi(parts[0])
	if err != nil {
		err := errors.New("invalid size range")
		logCtx.WithError(err).WithField("size", size).Error("Find by size")
		return err
	}

	sizeMax, err := strconv.Atoi(parts[1])
	if err != nil {
		err := errors.New("invalid size range")
		logCtx.WithError(err).WithField("size", size).Error("Find by size")
		return err
	}

	if sizeMin > sizeMax {
		err := errors.New("invalid size range")
		logCtx.WithError(err).WithField("size", size).Error("Find by size")
		return err
	}

	logCtx.WithFields(log.Fields{"min": sizeMin, "max": sizeMax}).Info("Find by size")

	sizeFunc := func(h map[string]int, row []string) (bool, error) {
		jobsReported := row[h["JobsReported"]]
		if jobsReported == "" {
			return false, nil
		}

		size, err := strconv.Atoi(jobsReported)
		if err != nil {
			return false, err
		}

		return size >= sizeMin && size <= sizeMax, nil
	}

	if err := readPPPFiles(ctx, logCtx, id, maxNum, sizeFunc); err != nil {
		return err
	}

	return nil
}
