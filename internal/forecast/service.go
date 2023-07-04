package forecast

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cicconee/weather-app/internal/app"
	"github.com/cicconee/weather-app/internal/geometry"
)

// ForecastAPI is the interface that wraps the GetGridpoint
// and GetHourlyForecast methods.
//
// GetGridpoint executes a HTTP GET request to the following url:
// https://api.weather.gov/points/{longitude},{latitude}
// It returns the server response in a GridpointAPIResource and
// any errors encountered.
//
// GetHourlyForecast executes a HTTP GET request to the following url:
// https://api.weather.gov/{grid_id}/{grid_x},{grid_y}/forecast/hourly
// It returns the server response in a HourlyAPIResource and any
// errors encountered.
type ForecastAPI interface {
	GetGridpoint(float64, float64) (GridpointAPIResource, error)
	GetHourlyForecast(string, int, int) (HourlyAPIResource, error)
}

// Service serves hourly forecasts. Hourly forecasts are retrieved from
// the NWS API.
//
// Service will store hourly forecasts in the database as they are requested.
// This allows Service to bypass unnecessary network calls to the NWS API.
// Only when new forecast areas are requested or if a forecast is out of date
// will service make network calls.
type Service struct {
	// The interface that will make network calls to the NWS API to get
	// the hourly forecasts.
	API ForecastAPI

	// The database connection.
	DB *sql.DB

	// The database storage.
	Store *Store
}

// New will return a pointer to a Service.
func New(api ForecastAPI, db *sql.DB) *Service {
	return &Service{
		API:   api,
		DB:    db,
		Store: NewStore(db),
	}
}

// Get will get the hourly forecast periods for the specified point.
func (s *Service) Get(ctx context.Context, point geometry.Point) (PeriodCollection, error) {
	gridpoint, err := s.Store.SelectGridpoint(ctx, point)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.write(ctx, point)
		}

		return PeriodCollection{}, fmt.Errorf("selecting gridpoint (point=%v): %w", point, err)
	}

	if time.Now().After(gridpoint.Timeline.ExpiresAt) {
		return s.update(ctx, gridpoint)
	}

	periodEntityCollection := PeriodEntityCollection{}
	if err := periodEntityCollection.Select(ctx, s.DB, gridpoint.ID); err != nil {
		return PeriodCollection{}, fmt.Errorf("selecting periods (gridpointID=%d): %w", gridpoint.ID, err)
	}

	location, err := time.LoadLocation(gridpoint.TimeZone)
	if err != nil {
		return PeriodCollection{}, fmt.Errorf("loading location (name=%s): %w", gridpoint.TimeZone, err)
	}

	periodCollection := periodEntityCollection.ToPeriods()
	periodCollection.loadTimeZone(location)
	periodCollection.Sort()

	return periodCollection, nil
}

// write will get the gridpoint and hourly forecast data from the NWS API. Once
// fetched, it will write the data to the database.
func (s *Service) write(ctx context.Context, point geometry.Point) (PeriodCollection, error) {
	gridpointResource, err := s.gridpoint(point)
	if err != nil {
		return PeriodCollection{}, fmt.Errorf("write: fetching gridpoint (lon=%f, lat=%f): %w", point.Lon(), point.Lat(), err)
	}

	// Some points are recognized by the NWS API as valid but do not have
	// a qualifying gridpoint associated with them. Most likely the point
	// resides some where in the ocean. The response from the NWS API would
	// be a 200 status code with GridID not set. These are points without
	// forecasts.
	if gridpointResource.GridID == "" {
		return PeriodCollection{}, app.NewServerResponseError(
			fmt.Errorf("write: no forecast for point (lon=%f, lat=%f)", point.Lon(), point.Lat()),
			fmt.Sprintf("%f,%f is not a supported area", point.Lon(), point.Lat()),
			http.StatusBadRequest)
	}

	hourlyResource, err := s.hourly(hourlyParams{
		GridID: gridpointResource.GridID,
		GridX:  gridpointResource.GridX,
		GridY:  gridpointResource.GridY,
	})
	if err != nil {
		return PeriodCollection{},
			fmt.Errorf("write: fetching hourly (GridID=%s, GridX=%d, GridY=%d): %w",
				gridpointResource.GridID,
				gridpointResource.GridX,
				gridpointResource.GridY,
				err)
	}

	gridpointEntity := gridpointResource.ToGridpointEntity()
	gridpointEntity.Geometry = hourlyResource.Geometry
	gridpointEntity.Timeline = hourlyResource.Timeline()
	if err := gridpointEntity.Insert(ctx, s.DB); err != nil {
		return PeriodCollection{}, fmt.Errorf("write: inserting gridpoint: %w", err)
	}

	periodEntityCollection := hourlyResource.ToPeriodEntityCollection()
	if err := periodEntityCollection.Insert(ctx, s.DB, gridpointEntity.ID); err != nil {
		return PeriodCollection{}, fmt.Errorf("write: inserting periods: %w", err)
	}

	return periodEntityCollection.ToPeriods(), nil
}

// update will get the hourly forecast data for a gridpoint from the NWS API. Once
// fetched, the gridpoint and hourly forecast will be updated in the database.
func (s *Service) update(ctx context.Context, gridpoint GridpointEntity) (PeriodCollection, error) {
	hourlyResource, err := s.hourly(hourlyParams{
		GridID: gridpoint.GridID,
		GridX:  gridpoint.GridX,
		GridY:  gridpoint.GridY,
	})
	if err != nil {
		return PeriodCollection{},
			fmt.Errorf("update: fetching hourly (GridID=%s, GridX=%d, GridY=%d): %w",
				gridpoint.GridID,
				gridpoint.GridX,
				gridpoint.GridY,
				err)
	}

	gridpoint.Timeline = hourlyResource.Timeline()
	if err := gridpoint.Update(ctx, s.DB); err != nil {
		return PeriodCollection{},
			fmt.Errorf("update: updating gridpoint (gridpoint.ID=%d): %w", gridpoint.ID, err)
	}

	periodEntityCollection := hourlyResource.ToPeriodEntityCollection()
	if err := periodEntityCollection.Update(ctx, s.DB, gridpoint.ID); err != nil {
		return PeriodCollection{},
			fmt.Errorf("update: updating periods (gridpoint.ID=%d): %w", gridpoint.ID, err)
	}

	return periodEntityCollection.ToPeriods(), nil
}

// gridpoint calls the GetGridpoint method of ForecastAPI for a point.
// If a 400 or 404 status code is returned it will return an Error with
// a safe message.
func (s *Service) gridpoint(point geometry.Point) (GridpointAPIResource, error) {
	gridpoint, err := s.API.GetGridpoint(point.Lon(), point.Lat())
	var apiErr *app.NWSAPIStatusCodeError
	switch {
	case err == nil:
		return gridpoint, nil
	case errors.As(err, &apiErr):
		if apiErr.StatusCode == 400 || apiErr.StatusCode == 404 {
			return GridpointAPIResource{}, app.NewServerResponseError(
				fmt.Errorf("not supported by api: %w", apiErr),
				fmt.Sprintf("%f,%f is not a supported area", point.Lon(), point.Lat()),
				http.StatusBadRequest)
		}

		return GridpointAPIResource{}, fmt.Errorf("unexpected status code: %w", apiErr)
	default:
		return GridpointAPIResource{}, err
	}
}

// hourlyParams is the parameters for the hourly method.
// When passing hourlyParams to hourly, all fields should be set.
type hourlyParams struct {
	GridID string
	GridX  int
	GridY  int
}

// hourly calls the GetHourlyForecast method of ForecastAPI for a gridpoint.
// If a 404 status code is returned it will return an Error with a safe message.
//
// It is a known issue that sometimes a 500 status code is returned from the NWS API
// hourly forecast endpoint for a valid gridpoint. The NWS API recommends retrying the
// request a few times. This will sometimes fix it.
func (s *Service) hourly(p hourlyParams) (HourlyAPIResource, error) {
	var (
		rErr     error
		attempts = 0
	)

	for attempts < 2 {
		hourly, err := s.API.GetHourlyForecast(p.GridID, p.GridX, p.GridY)
		var apiErr *app.NWSAPIStatusCodeError
		switch {
		case err == nil:
			return hourly, nil
		case errors.As(err, &apiErr):
			// If a valid gridpoint results in a 404 status code it is due to the
			// gridpoint being located in the ocean. The NWS API does not yet
			// support hourly forecasts for oceanic points.
			if apiErr.StatusCode == 404 {
				return HourlyAPIResource{}, app.NewServerResponseError(
					fmt.Errorf("not supported by api: %w", apiErr),
					"Oceanic points are not yet supported",
					http.StatusBadRequest)
			}

			// Set rErr incase this is the last attempt.
			if apiErr.StatusCode == 500 {
				rErr = app.NewServerResponseError(
					fmt.Errorf("not supported by api: %w", apiErr),
					"Not a supported area",
					http.StatusBadRequest)

				attempts++
			} else {
				return HourlyAPIResource{}, fmt.Errorf("unexpected status code: %w", apiErr)
			}
		default:
			return HourlyAPIResource{}, err
		}
	}

	return HourlyAPIResource{}, rErr
}
