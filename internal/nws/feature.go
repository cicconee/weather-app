package nws

import (
	"encoding/json"
	"fmt"
)

type feature struct {
	ID         string          `json:"id"`
	Properties json.RawMessage `json:"properties"`
}

func (f *feature) parseZone() (Zone, error) {
	var zone Zone
	if err := json.Unmarshal(f.Properties, &zone); err != nil {
		return zone, fmt.Errorf("failed unmarshalling *feature Properties field into Zone: %w", err)
	}

	zone.URI = f.ID

	return zone, nil
}

type featureCollection struct {
	Features []feature `json:"features"`
}
