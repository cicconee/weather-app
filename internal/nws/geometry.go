package nws

import (
	"encoding/json"
	"fmt"

	"github.com/cicconee/weather-app/internal/geometry"
)

type geo struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

func (g *geo) ParseMultiPolygon() (geometry.MultiPolygon, error) {
	var geo geometry.MultiPolygon
	var gErr error
	switch g.Type {
	case "":
		return geometry.MultiPolygon{}, nil
	case "Polygon":
		var polygon geometry.Polygon
		gErr = json.Unmarshal(g.Coordinates, &polygon)
		geo = polygon.AsMultiPolygon()
	case "MultiPolygon":
		var multiPolygon geometry.MultiPolygon
		gErr = json.Unmarshal(g.Coordinates, &multiPolygon)
		geo = multiPolygon
	default:
		return nil, fmt.Errorf("unsupported geometry type: %s", g.Type)
	}
	if gErr != nil {
		return nil, fmt.Errorf("failed parsing MultiPolygon (Type: %s): %w", g.Type, gErr)
	}

	return geo, nil
}

func (g *geo) ParsePolygon() (geometry.Polygon, error) {
	var geo geometry.Polygon
	var gErr error
	switch g.Type {
	case "":
		return geometry.Polygon{}, nil
	case "Polygon":
		var polygon geometry.Polygon
		gErr = json.Unmarshal(g.Coordinates, &polygon)
		geo = polygon
	default:
		return nil, fmt.Errorf("unsupported geometry type: %s", g.Type)
	}
	if gErr != nil {
		return nil, fmt.Errorf("failed parsing Polygon (Type: %s): %w", g.Type, gErr)
	}

	return geo, nil
}
