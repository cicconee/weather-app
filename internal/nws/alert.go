package nws

import (
	"time"

	"github.com/cicconee/weather-app/internal/geometry"
)

type Alert struct {
	URI           string
	ID            string           `json:"id"`
	AreaDesc      string           `json:"areaDesc"`
	AffectedZones []string         `json:"affectedZones"`
	References    []AlertReference `json:"references"`
	OnSet         time.Time        `json:"onset"`
	Expires       time.Time        `json:"expires"`
	Ends          time.Time        `json:"ends"`
	Status        string           `json:"status"`
	MessageType   string           `json:"messageType"`
	Category      string           `json:"category"`
	Severity      string           `json:"severity"`
	Certainty     string           `json:"certainty"`
	Urgency       string           `json:"urgency"`
	Event         string           `json:"event"`
	Headline      string           `json:"headline"`
	Description   string           `json:"description"`
	Instruction   string           `json:"instruction"`
	Response      string           `json:"response"`
	Geometry      geometry.Polygon
}

type AlertReference struct {
	ID string `json:"identifier"`
}

func (a *Alert) ReferenceIDs() []string {
	ss := []string{}
	for _, r := range a.References {
		ss = append(ss, r.ID)
	}
	return ss
}
