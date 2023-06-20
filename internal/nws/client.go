package nws

import (
	"encoding/json"
	"fmt"
	"net/http"
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

	var collection featureCollection
	if err := json.NewDecoder(res.Body).Decode(&collection); err != nil {
		return nil, fmt.Errorf("failed decoding http response: %w", err)
	}

	return &collection, nil
}

func (c *Client) GetZoneCollection(area string) ([]Zone, error) {
	collection, err := c.featureCollection(fmt.Sprintf("%s/zones?area=%s", API, area))
	if err != nil {
		return nil, err
	}

	var zoneCollection []Zone
	for _, f := range collection.Features {
		zone, err := f.parseZone()
		if err != nil {
			return nil, err
		}

		zoneCollection = append(zoneCollection, zone)
	}

	return zoneCollection, nil
}
