// Package common is common code for application main functions.
package common

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Service is the name of this service.
var Service string

// Init initializes servers.  This should be the first function called.
func Init(service string) {
	Service = service

	SetLogging("")
}

// SetInterrupt listens for SIGINT and SIGTERM events to cancel the server.
func SetInterrupt(ctx context.Context, cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
		case <-sigs:
			cancel()
		}
	}()
}
