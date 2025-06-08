package outbound

import (
	"context"
	"time"
)

type CarbonForecastPeriod struct {
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
	Forecast int       `json:"forecast"`
}

type CarbonIntensityPort interface {
	GetCarbonIntensity(ctx context.Context, from time.Time, to time.Time) ([]CarbonForecastPeriod, error)
}
