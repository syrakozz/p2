// Package cmd provides D1sTech console commands
package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "vox",
	Version: "1.0.0",
	Short:   "Vox client",
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "", false, "verbose")
	rootCmd.PersistentFlags().DurationP("timeout", "", time.Minute, "timeout")
}

// Execute the command line parser.
func Execute(ctx context.Context) {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SilenceErrors = true

	// ignore error
	cobra.CheckErr(rootCmd.ExecuteContext(ctx))
}
