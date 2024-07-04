// Service for all monday related functionality.
package main

import (
	"log"
	"net"

	"disruptive/cmd/common"
	"disruptive/config"
	"disruptive/rest/api"
	"disruptive/rest/firebase"
	"disruptive/rest/lib"
	"disruptive/rest/openai"
	"disruptive/rest/pinecone"
	"disruptive/rest/vox"
)

const (
	service = "sage"
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
		firebase.Routes,
		lib.Routes,
		openai.Routes,
		pinecone.Routes,
		vox.Routes,
	)
}
