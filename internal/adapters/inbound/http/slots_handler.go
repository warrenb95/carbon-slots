package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/warrenb95/carbon-slots/internal/ports/inbound"
)

type SlotsHandler struct {
	Controller     inbound.SlotController
	RequestTimeout time.Duration
}

func NewSlotsHandler(controller inbound.SlotController, timeoutDuration time.Duration) *SlotsHandler {
	return &SlotsHandler{
		Controller:     controller,
		RequestTimeout: timeoutDuration,
	}
}

func (h *SlotsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request for %s", r.URL.Path)

	// Parse query params
	duration := r.URL.Query().Get("duration")
	slotDuration, err := strconv.ParseInt(duration, 10, 64)
	if err != nil {
		slotDuration = 30 // Default to 30 minutes if parsing fails
	}
	if slotDuration > 1440 || slotDuration < 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid duration, must be between 0 and 1440 minutes")
		log.Printf("Invalid duration: %d minutes", slotDuration)
		return
	}

	contineousStr := r.URL.Query().Get("contineous")
	contineousStr = strings.TrimSpace(contineousStr)
	contineousStr = strings.ToLower(contineousStr)
	var contineous bool
	if contineousStr == "true" {
		contineous = true
	}

	ctxTimeout, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	slots, err := h.Controller.FindSlots(ctxTimeout, time.Duration(slotDuration)*time.Minute, contineous)
	if err != nil {
		// TODO: add error for no results
		log.Printf("Error finding slots: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to find slots")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slots)
}
