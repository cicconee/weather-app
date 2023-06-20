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
