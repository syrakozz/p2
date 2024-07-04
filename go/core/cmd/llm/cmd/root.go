// Package cmd provides D1sTech console commands
package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "llm",
	Version: "1.0.0",
	Short:   "LLM client",
}

func init() {
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose")
}

// Execute the command line parser.
func Execute(ctx context.Context) {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SilenceErrors = true

	// ignore error
	cobra.CheckErr(rootCmd.ExecuteContext(ctx))
}
