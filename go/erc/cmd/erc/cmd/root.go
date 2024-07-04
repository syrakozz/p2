// Package cmd provides erc console commands
package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "erc",
	Version: "1.0.0",
	Short:   "erc client",
}

// Execute the command line parser.
func Execute(ctx context.Context) {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SilenceErrors = true

	// ignore error
	cobra.CheckErr(rootCmd.ExecuteContext(ctx))
}
