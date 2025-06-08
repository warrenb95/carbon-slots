package carbonintensityapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/warrenb95/carbon-slots/internal/ports/outbound"
)

// Adapter implements the outbound.CarbonIntensityPort interface to fetch carbon intensity data from an external API.
type Adapter struct {
	BaseURL string
	Client  *http.Client
}

// NewAdapter creates a new Adapter instance with the specified base URL for the carbon intensity API.
func NewAdapter(baseURL string) *Adapter {
	return &Adapter{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

type apiResponse struct {
	Data []struct {
		From      string `json:"from"`
		To        string `json:"to"`
		Intensity struct {
			Forecast int `json:"forecast"`
		} `json:"intensity"`
	} `json:"data"`
}

// GetCarbonIntensity fetches carbon intensity data for the specified time range.
func (a *Adapter) GetCarbonIntensity(ctx context.Context, from, to time.Time) ([]outbound.CarbonForecastPeriod, error) {
	url := fmt.Sprintf("%s/intensity/%s/fw24h", a.BaseURL, from.Format("2006-01-02T15:04Z"))

	operation := func() (*apiResponse, error) {
		resp, err := a.Client.Get(url)
		if err != nil {
			log.Printf("carbon intensity get request error: %v", err)
			return nil, fmt.Errorf("carbon intensity get request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusInternalServerError {
			log.Printf("carbon intensity API returned internal server error: %s", resp.Status)
			return nil, fmt.Errorf("carbon intensity api error: %s", resp.Status)
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("carbon intensity API returned unexpected status code: %d", resp.StatusCode)
			return nil, backoff.Permanent(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
		}

		var apiResp apiResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			log.Printf("error decoding carbon intensity response: %v", err)
			return nil, err
		}
		return &apiResp, nil
	}

	apiResp, err := backoff.Retry(ctx, operation, backoff.WithBackOff(backoff.NewExponentialBackOff()), backoff.WithMaxTries(3))
	if err != nil {
		log.Printf("failed to get carbon intensity data: %v", err)
		return nil, fmt.Errorf("failed to get carbon intensity data: %w", err)
	}

	if len(apiResp.Data) == 0 {
		log.Println("no data in carbon intensity response")
		return nil, errors.New("no data in carbon intensity response")
	}

	resp := make([]outbound.CarbonForecastPeriod, len(apiResp.Data))
	for i, data := range apiResp.Data {
		fromTime, err := time.Parse("2006-01-02T15:04Z", data.From)
		if err != nil {
			log.Printf("invalid from time format: %v", err)
			return nil, fmt.Errorf("invalid from time format: %w", err)
		}
		toTime, err := time.Parse("2006-01-02T15:04Z", data.To)
		if err != nil {
			log.Printf("invalid to time format: %v", err)
			return nil, fmt.Errorf("invalid to time format: %w", err)
		}

		resp[i] = outbound.CarbonForecastPeriod{
			From:     fromTime,
			To:       toTime,
			Forecast: data.Intensity.Forecast,
		}
	}

	slices.SortFunc(resp, func(a, b outbound.CarbonForecastPeriod) int {
		return a.From.Compare(b.From)
	})

	return resp, nil
}
