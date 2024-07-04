// Application to manage ERC and PPP data.
package main

import (
	"context"

	log "github.com/sirupsen/logrus"

	"disruptive/cmd/common"
	"disruptive/cmd/erc/cmd"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	log.AddHook(&varsHook{})

	common.SetInterrupt(ctx, cancel)

	cmd.Execute(ctx)
}

type varsHook struct {
}

func (*varsHook) Levels() []log.Level {
	return []log.Level{log.InfoLevel, log.WarnLevel}
}

func (*varsHook) Fire(entry *log.Entry) error {
	delete(entry.Data, "fid")
	return nil
}
