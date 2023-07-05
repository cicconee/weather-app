package nws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cicconee/weather-app/internal/app"
	"github.com/cicconee/weather-app/internal/forecast"
)

const API = "https://api.weather.gov"

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	HTTP      HTTPDoer
	UserAgent string
}

var DefaultClient = &Client{
	HTTP: defaultHTTP(),
}

func (c *Client) http() HTTPDoer {
	if c.HTTP == nil {
		return DefaultClient.HTTP
	}

	return c.HTTP
}

func (c *Client) get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating GET request: %w", err)
	}

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	res, err := c.http().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute GET request: %w", err)
	}

	return res, nil
}

func (c *Client) featureCollection(url string) (*featureCollection, error) {
	res, err := c.get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to getting http response: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var statusErr *app.NWSAPIStatusCodeError
		if err := json.NewDecoder(res.Body).Decode(&statusErr); err != nil {
			statusErr = &app.NWSAPIStatusCodeError{StatusCode: res.StatusCode}
			return nil, fmt.Errorf("%w: failed to decode app.NWSAPIStatusCodeError Detail field: %v", statusErr, err)
		}
		return nil, statusErr
	}

	var collection featureCollection
	if err := json.NewDecoder(res.Body).Decode(&collection); err != nil {
		return nil, fmt.Errorf("failed decoding http response: %w", err)
	}

	return &collection, nil
}

func (c *Client) feature(url string) (*feature, error) {
	res, err := c.get(url)
	if err != nil {
		return nil, fmt.Errorf("failed getting http response: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var statusErr *app.NWSAPIStatusCodeError
		if err := json.NewDecoder(res.Body).Decode(&statusErr); err != nil {
			statusErr = &app.NWSAPIStatusCodeError{StatusCode: res.StatusCode}
			return nil, fmt.Errorf("%w: failed to decode app.NWSAPIStatusCodeError Detail field: %v", statusErr, err)
		}

		return nil, statusErr
	}

	var f feature
	if err := json.NewDecoder(res.Body).Decode(&f); err != nil {
		return nil, fmt.Errorf("failed decoding http response: %w", err)
	}

	return &f, nil
}

func (c *Client) GetZoneCollection(area string) ([]Zone, error) {
	collection, err := c.featureCollection(fmt.Sprintf("%s/zones?area=%s", API, area))
	if err != nil {
		return nil, fmt.Errorf("failed to get feature collection: %w", err)
	}

	var zoneCollection []Zone
	for _, f := range collection.Features {
		zone, err := f.parseZone()
		if err != nil {
			return nil, fmt.Errorf("failed to parse Zone (URI: %s): %w", f.ID, err)
		}

		zoneCollection = append(zoneCollection, zone)
	}

	return zoneCollection, nil
}

func (c *Client) GetZone(zoneType string, zoneCode string) (Zone, error) {
	feat, err := c.feature(fmt.Sprintf("%s/zones/%s/%s", API, zoneType, zoneCode))
	if err != nil {
		return Zone{}, fmt.Errorf("failed to get feature: %w", err)
	}

	zone, err := feat.parseZone()
	if err != nil {
		return Zone{}, fmt.Errorf("failed to parse Zone: %w", err)
	}

	return zone, nil
}

func (c *Client) GetActiveAlerts(states ...string) ([]Alert, error) {
	if len(states) == 0 {
		return []Alert{}, nil
	}

	collection, err := c.featureCollection(
		fmt.Sprintf("%s/alerts/active?status=actual&area=%s",
			API,
			strings.Join(states, ",")))
	if err != nil {
		return nil, fmt.Errorf("failed to get feature collection: %w", err)
	}

	var alerts []Alert
	for _, f := range collection.Features {
		var alert Alert
		if err := json.Unmarshal(f.Properties, &alert); err != nil {
			return nil, fmt.Errorf("failed to unmarshal alert properties: %w", err)
		}

		geo, err := f.Geometry.ParsePolygon()
		if err != nil {
			return nil, fmt.Errorf("failed to parse Geometry as a Polygon: %w", err)
		}

		alert.URI = f.ID
		alert.Geometry = geo
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (c *Client) GetGridpoint(x, y float64) (forecast.GridpointAPIResource, error) {
	feature, err := c.feature(fmt.Sprintf("%s/points/%f,%f", API, x, y))
	if err != nil {
		return forecast.GridpointAPIResource{}, err
	}

	gridpoint := forecast.GridpointAPIResource{}
	if err := json.Unmarshal(feature.Properties, &gridpoint); err != nil {
		return forecast.GridpointAPIResource{}, fmt.Errorf("parsing gridpoint: %w", err)
	}

	return gridpoint, nil
}

func (c *Client) GetHourlyForecast(id string, x, y int) (forecast.HourlyAPIResource, error) {
	feature, err := c.feature(fmt.Sprintf("%s/gridpoints/%s/%d,%d/forecast/hourly?units=us",
		API, id, x, y))
	if err != nil {
		return forecast.HourlyAPIResource{}, err
	}

	hourly := forecast.HourlyAPIResource{}
	if err := json.Unmarshal(feature.Properties, &hourly); err != nil {
		return forecast.HourlyAPIResource{}, fmt.Errorf("nws: failed to parse forecast.Hourly: %w", err)
	}

	polygon, err := feature.Geometry.ParsePolygon()
	if err != nil {
		return forecast.HourlyAPIResource{}, fmt.Errorf("nws: failed to parse forecast geometry: %w", err)
	}

	hourly.Geometry = polygon

	return hourly, nil
}
