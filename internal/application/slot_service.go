package application

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/warrenb95/carbon-slots/internal/domain"
	"github.com/warrenb95/carbon-slots/internal/ports/outbound"
)

type SlotService struct {
	CarbonAPI outbound.CarbonIntensityPort
}

func NewSlotService(carbonAPI outbound.CarbonIntensityPort) *SlotService {
	return &SlotService{
		CarbonAPI: carbonAPI,
	}
}

// FindSlots finds Slots for the given time range
func (s *SlotService) FindSlots(ctx context.Context, duration time.Duration, contineous bool) ([]domain.Slot, error) {
	startTime := time.Now().UTC()
	endTime := startTime.Add(24 * time.Hour)

	periods, err := s.CarbonAPI.GetCarbonIntensity(ctx, startTime, endTime)
	if err != nil {
		log.Printf("Error fetching carbon intensity data: %v", err)
		return nil, err
	}

	if len(periods) == 0 {
		log.Println("No carbon intensity data available for the requested period")
		return nil, fmt.Errorf("no carbon intensity data available for the requested period")
	}

	if contineous {
		// Slide a window of 'duration' over the periods, return the first window that fits
		for start := periods[0].From.UTC(); start.Add(duration).Before(endTime) || start.Add(duration).Equal(endTime); start = start.Add(15 * time.Minute) {
			end := start.Add(duration)
			avg, err := WeightedAverage(periods, start, end)
			if err != nil {
				continue
			}
			return []domain.Slot{{
				ValidFrom: start,
				ValidTo:   end,
				Carbon:    domain.Carbon{Intensity: avg},
			}}, nil
		}

		log.Println("No continuous slot found for the requested duration")
		return nil, fmt.Errorf("no continuous slot found")
	}

	// Non-continuous: Patch together lowest-intensity periods to sum to duration
	// 1. Sort periods by intensity ascending
	sort.Slice(periods, func(i, j int) bool {
		return periods[i].Forecast < periods[j].Forecast
	})

	var (
		slots        []domain.Slot
		accumulated  time.Duration
		remainingDur = duration
	)

	for _, p := range periods {
		periodDur := p.To.Sub(p.From)
		useDur := min(remainingDur, periodDur)
		// Weighted average for this slot (may be partial period)
		slotAvg := p.Forecast
		if useDur < periodDur {
			// Partial period, weighted average is just the period's forecast
			// (since it's uniform within the period)
			// Start is p.From, end is p.From + useDur
			slots = append(slots, domain.Slot{
				ValidFrom: p.From,
				ValidTo:   p.From.Add(useDur),
				Carbon:    domain.Carbon{Intensity: slotAvg},
			})
		} else {
			slots = append(slots, domain.Slot{
				ValidFrom: p.From,
				ValidTo:   p.To,
				Carbon:    domain.Carbon{Intensity: slotAvg},
			})
		}
		accumulated += useDur
		remainingDur -= useDur
		if accumulated >= duration {
			break
		}
	}

	if accumulated < duration {
		log.Println("Not enough periods to cover the requested duration")
		return nil, fmt.Errorf("not enough periods to cover requested duration")
	}

	return slots, nil
}

// WeightedAverage calculates the weighted average forecast for the slot window.
// It supports partial periods.
func WeightedAverage(periods []outbound.CarbonForecastPeriod, from, to time.Time) (int, error) {
	var totalWeight float64
	var weightedSum float64

	for _, p := range periods {
		overlapStart := maxTime(from, p.From)
		overlapEnd := minTime(to, p.To)
		if overlapStart.Before(overlapEnd) {
			weight := overlapEnd.Sub(overlapStart).Minutes()
			weightedSum += float64(p.Forecast) * weight
			totalWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0, fmt.Errorf("no overlap between periods and slot window")
	}

	return int(weightedSum / totalWeight), nil
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
