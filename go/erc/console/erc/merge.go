package erc

import (
	"context"
)

// Merge PPP and RocketReach records.
func Merge(ctx context.Context, basename string) error {
	pppHeader, pppHeaderIndexes, pppRowsByLoanNumber, err := mergePpp(ctx, basename)
	if err != nil {
		return err
	}

	loansProcessed, err := mergeRocketReach(ctx, basename, pppHeaderIndexes, pppRowsByLoanNumber)
	if err != nil {
		return err
	}

	if err := mergeRocketReachNP(ctx, basename, pppHeader, pppHeaderIndexes, pppRowsByLoanNumber, loansProcessed); err != nil {
		return err
	}

	return nil
}
