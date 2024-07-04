// Package products manages dt products functions.
package products

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"disruptive/lib/common"
	"disruptive/lib/firestore"
)

// ImportProductDevice generates and imports product device mac addresses to firestore.
func ImportProductDevice(ctx context.Context, productName string, macAddress uint64, numDevices int, force bool) error {
	logCtx := slog.With("fid", "console.products.ImportProductDevice")

	start := time.Now()

	path := fmt.Sprintf("products/%s/ids", productName)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("product collection not found")
		return common.ErrNotFound{}
	}

	doc := map[string]any{}
	skipped := []string{}
	valid := 0

	for i := 0; i < numDevices; i++ {
		mac, err := net.ParseMAC(common.IntToMACAddress(macAddress + uint64(i)))
		if err != nil {
			skipped = append(skipped, mac.String())
			continue
		}

		if force {
			if _, err := collection.Doc(mac.String()).Set(ctx, doc); err != nil {
				err = common.ConvertGRPCError(err)
				logCtx.Error("unable to set device document", "error", err)

				skipped = append(skipped, mac.String())
				continue
			}

			valid++
		} else {
			if _, err := collection.Doc(mac.String()).Create(ctx, doc); err != nil {
				err = common.ConvertGRPCError(err)
				if errors.Is(err, common.ErrAlreadyExists{}) {
					fmt.Println(mac.String() + " already exists. Skipping.")
					skipped = append(skipped, mac.String())
					continue
				}

				logCtx.Error("unable to create device document", "error", err)
				skipped = append(skipped, mac.String())
				continue
			}

			valid++
		}
	}

	fmt.Println()
	fmt.Println(fmt.Sprintf("Total devices added: %d out of %d", valid, numDevices))
	fmt.Println()
	fmt.Println("Total skipped devices: ", len(skipped))
	fmt.Println("Skipped MAC addresses: ", skipped)
	fmt.Println()
	fmt.Println("Elapsed time: ", time.Since(start))
	fmt.Println()

	return nil
}
