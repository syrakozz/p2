// Service for all monday related functionality.
package main

import (
	"log"
	"net"

	"disruptive/cmd/common"
	"disruptive/config"
	"disruptive/rest/api"
	"disruptive/rest/monday"
)

const (
	service = "monday"
)

func main() {
	common.Init(service)

	l, err := net.Listen("tcp", ":"+config.VARS.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	common.StartHTTPServer(
		l,
		api.Routes,
		monday.Routes,
	)
}
