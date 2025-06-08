package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warrenb95/carbon-slots/internal/application"
	"github.com/warrenb95/carbon-slots/internal/application/mocks"
	"github.com/warrenb95/carbon-slots/internal/domain"
	"github.com/warrenb95/carbon-slots/internal/ports/outbound"
)

//go:generate moq -pkg mocks -out mocks/mock_carbon_api.go ../ports/outbound CarbonIntensityPort

func TestFindSlots(t *testing.T) {
	now := time.Now().Add(time.Minute * 30).UTC()
	periods := []outbound.CarbonForecastPeriod{
		{From: now, To: now.Add(30 * time.Minute), Forecast: 100},
		{From: now.Add(30 * time.Minute), To: now.Add(60 * time.Minute), Forecast: 200},
		{From: now.Add(60 * time.Minute), To: now.Add(90 * time.Minute), Forecast: 50},
	}

	tests := map[string]struct {
		periods    []outbound.CarbonForecastPeriod
		err        error
		duration   time.Duration
		continuous bool
		wantSlots  []domain.Slot
		wantErr    bool
	}{
		"continuous slot found": {
			periods:    periods,
			duration:   30 * time.Minute,
			continuous: true,
			wantSlots: []domain.Slot{{
				ValidFrom: now,
				ValidTo:   now.Add(30 * time.Minute),
				Carbon:    domain.Carbon{Intensity: 100},
			}},
			wantErr: false,
		},
		"continuous slot not found": {
			periods:    []outbound.CarbonForecastPeriod{},
			duration:   30 * time.Minute,
			continuous: true,
			wantSlots:  nil,
			wantErr:    true,
		},
		"non-continuous, exact periods": {
			periods:    periods,
			duration:   60 * time.Minute,
			continuous: false,
			wantSlots: []domain.Slot{
				{ValidFrom: now.Add(60 * time.Minute), ValidTo: now.Add(90 * time.Minute), Carbon: domain.Carbon{Intensity: 50}},
				{ValidFrom: now, ValidTo: now.Add(30 * time.Minute), Carbon: domain.Carbon{Intensity: 100}},
			},
			wantErr: false,
		},
		"non-continuous, partial period": {
			periods:    periods,
			duration:   45 * time.Minute,
			continuous: false,
			wantSlots: []domain.Slot{
				{ValidFrom: now.Add(60 * time.Minute), ValidTo: now.Add(90 * time.Minute), Carbon: domain.Carbon{Intensity: 50}},
				{ValidFrom: now, ValidTo: now.Add(15 * time.Minute), Carbon: domain.Carbon{Intensity: 100}},
			},
			wantErr: false,
		},
		"api error": {
			periods:    nil,
			err:        errors.New("api error"),
			duration:   30 * time.Minute,
			continuous: false,
			wantSlots:  nil,
			wantErr:    true,
		},
		"not enough periods": {
			periods:    periods[:1],
			duration:   60 * time.Minute,
			continuous: false,
			wantSlots:  nil,
			wantErr:    true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			carbonPortMock := &mocks.CarbonIntensityPortMock{
				GetCarbonIntensityFunc: func(ctx context.Context, from, to time.Time) ([]outbound.CarbonForecastPeriod, error) {
					return tc.periods, tc.err
				},
			}

			service := application.NewSlotService(carbonPortMock)
			slots, err := service.FindSlots(context.Background(), tc.duration, tc.continuous)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantSlots, slots)
		})
	}
}

func TestWeightedAverage(t *testing.T) {
	now := time.Date(2025, 1, 9, 1, 0, 0, 0, time.UTC)
	periods := []outbound.CarbonForecastPeriod{
		{From: now, To: now.Add(30 * time.Minute), Forecast: 100},
		{From: now.Add(30 * time.Minute), To: now.Add(60 * time.Minute), Forecast: 200},
	}
	tests := []struct {
		name     string
		from     time.Time
		to       time.Time
		expected int
		wantErr  bool
	}{
		{
			name:     "full overlap first period",
			from:     now,
			to:       now.Add(30 * time.Minute),
			expected: 100,
			wantErr:  false,
		},
		{
			name:     "partial overlap both periods",
			from:     now.Add(15 * time.Minute),
			to:       now.Add(45 * time.Minute),
			expected: 150, // (100*15 + 200*15)/30 = 150
			wantErr:  false,
		},
		{
			name:    "no overlap",
			from:    now.Add(2 * time.Hour),
			to:      now.Add(3 * time.Hour),
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := application.WeightedAverage(periods, tc.from, tc.to)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}
