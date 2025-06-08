package main

import (
	"log"

	"github.com/warrenb95/carbon-slots/internal/adapters/inbound/http"
)

func main() {
	server := http.NewServer(":3000", "https://api.carbonintensity.org.uk")
	if err := server.Start(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
