package state

import "time"

type SaveResult struct {
	State     string
	Writes    []SaveZoneResult
	Fails     []SaveZoneFailure
	CreatedAt time.Time
}

func (s *SaveResult) TotalZones() int {
	return len(s.Writes) + len(s.Fails)
}

type SaveZoneResult struct {
	URI  string
	Code string
	Type string
}

type SaveZoneFailure struct {
	SaveZoneResult
	err error
}
