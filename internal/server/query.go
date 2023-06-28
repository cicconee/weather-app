package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cicconee/weather-app/internal/geometry"
)

type QueryParameterError struct {
	Msg string
	error
}

func (p *QueryParameterError) ServerErrorResponse() (int, string) {
	return http.StatusBadRequest, p.Msg
}

// ParsePoint takes longitude and latitude as
// strings (lonStr, latStr) and returns them
// as a geometry.Point.
//
// If parsing fails an error is returned as a
// QueryParameterError.
func ParsePoint(lonStr string, latStr string) (geometry.Point, error) {
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		qErr := &QueryParameterError{
			Msg:   "Invalid longitude",
			error: fmt.Errorf("failed to parse lon: %w", err),
		}
		return geometry.Point{}, qErr
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		qErr := &QueryParameterError{
			Msg:   "Invalid latitude",
			error: fmt.Errorf("failed to parse lot: %w", err)}
		return geometry.Point{}, qErr
	}

	return geometry.NewPoint(lon, lat), nil
}
