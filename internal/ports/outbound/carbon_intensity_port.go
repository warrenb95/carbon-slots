package outbound

import (
	"context"
	"time"
)

// CarbonForecastPeriod represents a period of time with a carbon intensity forecast.
type CarbonForecastPeriod struct {
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
	Forecast int       `json:"forecast"`
}

// CarbonIntensityPort defines the interface for fetching carbon intensity data.
type CarbonIntensityPort interface {
	GetCarbonIntensity(ctx context.Context, from time.Time, to time.Time) ([]CarbonForecastPeriod, error)
}
