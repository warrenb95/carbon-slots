package inbound

import (
	"context"
	"time"

	"github.com/warrenb95/carbon-slots/internal/domain"
)

type SlotController interface {
	// FindSlots finds Slots for the given time range
	FindSlots(ctx context.Context, duration time.Duration, contineous bool) ([]domain.Slot, error)
}
