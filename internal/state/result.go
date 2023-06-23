package state

import "time"

type SaveResult struct {
	State     string
	Writes    []Zone
	Fails     []SaveZoneFailure
	CreatedAt time.Time
}

func (s *SaveResult) TotalZones() int {
	return len(s.Writes) + len(s.Fails)
}

type SaveZoneResult struct {
	Writes []Zone
	Fails  []SaveZoneFailure
}

type SaveZoneFailure struct {
	URI  string
	Code string
	Type string
	err  error
}
