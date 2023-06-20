package nws

import "time"

type Zone struct {
	URI           string
	Code          string    `json:"id"`
	Type          string    `json:"type"`
	Name          string    `json:"name"`
	EffectiveDate time.Time `json:"effectiveDate"`
	State         string    `json:"state"`
}
