package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"disruptive/config"
	"disruptive/console/erc"
)

var pppCmd = &cobra.Command{
	Use:   "ppp",
	Short: "ppp commands",
	Long:  "ppp commands.",
}

var uploadPppCmd = &cobra.Command{
	Use:   "upload [file]",
	Short: "upload ppp csv file",
	Long:  "Upload a ppp csv file.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(args[0]); os.IsNotExist(err) {
			return errors.New("file does not exist")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		erc.Upload(cmd.Root().Context(), args[0])
	},
}

var findPppCmd = &cobra.Command{
	Use:   "find [num]",
	Short: "find records",
	Long:  "Find records.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if _, err := strconv.Atoi(args[0]); err != nil {
			return errors.New("number of records must be a number")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		num, _ := strconv.Atoi(args[0])

		size, _ := cmd.Flags().GetString("size")

		if err := erc.Find(cmd.Root().Context(), num, size); err != nil {
			log.WithError(err).Error("unable to run find")
			os.Exit(1)
		}
	},
}

var mergePppCmd = &cobra.Command{
	Use:   "merge [basename]",
	Short: "merge records",
	Long:  "Merge records.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {

		info, err := os.Stat(filepath.Join(config.VARS.ERCDataRoot, "process", args[0]+"-rocketreach-out.csv"))
		if os.IsNotExist(err) || info.IsDir() {
			return fmt.Errorf("file not found: %s-rocketreach-out.csv", args[0])
		}

		info, err = os.Stat(filepath.Join(config.VARS.ERCDataRoot, "process", args[0]+"-ppp.csv"))
		if os.IsNotExist(err) || info.IsDir() {
			return fmt.Errorf("file not found: %s-ppp.csv", args[0])
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := erc.Merge(cmd.Root().Context(), args[0]); err != nil {
			log.WithError(err).Error("unable to run merge")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(pppCmd)

	pppCmd.AddCommand(uploadPppCmd)

	pppCmd.AddCommand(findPppCmd)
	findPppCmd.Flags().StringP("size", "s", "", "size range: (min-max)")

	pppCmd.AddCommand(mergePppCmd)
}
