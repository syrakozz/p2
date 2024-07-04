package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"disruptive/cmd/common"
	"disruptive/console/csv"
)

var csvCmd = &cobra.Command{
	Use:   "csv",
	Short: "csv commands",
	Long:  "csv commands.",
}

var splitCsvCmd = &cobra.Command{
	Use:   "split [num] [file]",
	Short: "split a csv file into [num] files",
	Long:  "Split a csv file into [num] files preserving the header.",
	Args:  cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if _, err := strconv.Atoi(args[0]); err != nil {
			return errors.New("[num] must be a number")
		}

		if !common.FileExistsValidator(args[1]) {
			return fmt.Errorf("%q doesn't exist", args[1])
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		num, _ := strconv.Atoi(args[0])

		if err := csv.SplitCSV(cmd.Root().Context(), num, args[1]); err != nil {
			log.WithError(err).Error("unable to split csv file")
			os.Exit(1)
		}

	},
}

func init() {
	rootCmd.AddCommand(csvCmd)

	csvCmd.AddCommand(splitCsvCmd)
}
