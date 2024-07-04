package erc

import (
	"context"

	log "github.com/sirupsen/logrus"

	e "disruptive/lib/erc"
)

// Naics displays the business establishment value for an NAICS code
func Naics(ctx context.Context, code string) {
	v := e.Naics[code]
	if v == "" {
		log.WithFields(log.Fields{"code": code, "establishment": "Unknown"}).Warn("NAICS")
		return
	}
	log.WithFields(log.Fields{"code": code, "establishment": v}).Info("NAICS")
}
