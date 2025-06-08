package http

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/warrenb95/carbon-slots/internal/adapters/outbound/carbonintensityapi"
	"github.com/warrenb95/carbon-slots/internal/application"
)

type Server struct {
	HTTPServer *http.Server
}

func NewServer(addr string, carbonAPIBaseURL string) *Server {
	// Outbound adapter
	carbonAdapter := carbonintensityapi.NewAdapter(carbonAPIBaseURL)

	// Application service
	slotService := application.NewSlotService(carbonAdapter)

	// Inbound adapter (handler)
	// Default timeout for slots is set to 10 seconds.
	// TODO: should update to use env var for the timeout.
	slotsHandler := NewSlotsHandler(slotService, 10*time.Second)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/slots", slotsHandler)

	return &Server{
		HTTPServer: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.HTTPServer.Addr)
	return s.HTTPServer.ListenAndServe()
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: msg})
}
