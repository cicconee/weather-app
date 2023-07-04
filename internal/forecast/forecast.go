package forecast

import (
	"time"

	"github.com/cicconee/weather-app/internal/geometry"
)

// HourlyAPIResource is the hourly forecast data that is returned by ForecastAPI.
// HourlyAPIResource should never be explicitly created and only be used when it
// is returned by ForecastAPI.
//
// A hourly forecast is divided into 1-hour periods for a specific geological area.
//
// The ToPeriodEntityCollection method will return each PeriodAPIResource as a PeriodEntity.
// The Timeline method will create the Timeline for a hourly forecast.
type HourlyAPIResource struct {
	// The geometric boundary that this hourly forecast is valid for. All coordinates
	// residing in this boundary will use this forecast.
	Geometry geometry.Polygon

	// The time this data was generated at on the NWS API server.
	GeneratedAt time.Time `json:"generatedAt"`

	// The forecast periods. Each PeriodAPIResource holds weather information for a 1-hour
	// period.
	Periods []PeriodAPIResource `json:"periods"`
}

// ToPeriodEntityCollection returns Periods as a PeriodEntityCollection.
func (h *HourlyAPIResource) ToPeriodEntityCollection() PeriodEntityCollection {
	periods := PeriodEntityCollection{}
	for _, p := range h.Periods {
		periods = append(periods, p.ToPeriodEntity())
	}
	return periods
}

// Timeline returns this HourlyAPIResource GeneratedAt time and
// when it will expire as a Timeline. Both times are in UTC format.
func (h *HourlyAPIResource) Timeline() Timeline {
	return Timeline{
		GeneratedAt: h.GeneratedAt.UTC(),
		ExpiresAt:   h.GeneratedAt.Add(time.Hour).UTC(),
	}
}

// Timeline is the times forecast data was generated at and when it
// will be expired.
type Timeline struct {
	// The creation time of the forecast data.
	GeneratedAt time.Time

	// The expiration time of the forecast data.
	ExpiresAt time.Time
}
