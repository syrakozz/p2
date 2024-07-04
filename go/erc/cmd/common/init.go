// Package common is common code for application main functions.
package common

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"disruptive/config"
)

// Service is the name of this service.
var Service string

// Init initializes servers.  This should be the first function called.
func Init(service string) {
	Service = service

	log.SetOutput(&StdLogWriter{})

	switch config.VARS.LoggingLevel {
	case log.TraceLevel.String():
		// Something very low level.
		log.SetLevel(log.TraceLevel)
	case log.DebugLevel.String():
		// Useful debugging information.
		log.SetLevel(log.DebugLevel)
	case log.WarnLevel.String():
		// You should probably take a look at this.
		log.SetLevel(log.WarnLevel)
	case log.ErrorLevel.String():
		// Something failed but I'm not quitting.
		log.SetLevel(log.ErrorLevel)
	case log.FatalLevel.String():
		// Bye.	Calls os.Exit(1) after logging.
		log.SetLevel(log.FatalLevel)
	case log.PanicLevel.String():
		// Bailing. Calls panic() after logging.
		log.SetLevel(log.PanicLevel)
	default:
		// Something noteworthy happened.
		log.SetLevel(log.InfoLevel)
	}

	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	log.AddHook(NewExtraLogVars(service, config.VARS.Env))

	if config.VARS.Env != "local" {
		log.WithFields(log.Fields{
			"image":    config.VARS.BuildImage,
			"commit":   config.VARS.BuildCommit,
			"datetime": config.VARS.BuildDateTime,
		}).Infof("build")
	}
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
