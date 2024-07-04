package cmd

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"disruptive/console/products"
)

var productsCmd = &cobra.Command{
	Use:   "products",
	Short: "products commands",
	Long:  "products commands.",
}

var importProductDeviceCmd = &cobra.Command{
	Use:   "import [product] [device_mac_address] [num_devices]",
	Short: "import commands",
	Long:  "Import config commands.",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		product := args[0]

		macAddress, err := strconv.ParseUint(args[1], 16, 64)
		if err != nil {
			os.Exit(400)
		}

		numDevices, err := strconv.Atoi(args[2])
		if err != nil {
			os.Exit(400)
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			os.Exit(400)
		}

		if err := products.ImportProductDevice(cmd.Root().Context(), product, macAddress, numDevices, force); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(productsCmd)

	productsCmd.AddCommand(importProductDeviceCmd)
	importProductDeviceCmd.Flags().Bool("force", false, "force existing devices to be overwritten")
}
