// Application to manage admin D1sTech functionality
package main

import (
	"context"

	"disruptive/cmd/common"
	"disruptive/cmd/dt/cmd"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	common.SetInterrupt(ctx, cancel)

	cmd.Execute(ctx)
}
