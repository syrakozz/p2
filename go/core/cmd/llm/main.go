// Application to manage LLM functionality
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"disruptive/cmd/common"
	"disruptive/cmd/llm/cmd"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	common.SetLogging("")
	setInterrupt(ctx, cancel)

	cmd.Execute(ctx)
}

func setInterrupt(ctx context.Context, cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGQUIT) // Ctrl-\
	go func() {
		select {
		case <-ctx.Done():
		case <-sigs:
			cancel()
		}
	}()
}
