package inbound

import (
	"context"
	"time"

	"github.com/warrenb95/carbon-slots/internal/domain"
)

// SlotController defines the interface for finding carbon slots.
type SlotController interface {
	// FindSlots finds Slots for the given time range
	FindSlots(ctx context.Context, duration time.Duration, contineous bool) ([]domain.Slot, error)
}
