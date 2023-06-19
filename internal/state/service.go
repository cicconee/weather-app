package state

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Service struct{}

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

func (s *Service) Save(ctx context.Context, stateID string) (SaveResult, error) {
	stateID = strings.ToUpper(stateID)

	return SaveResult{
		State: stateID,
		Writes: []SaveZoneResult{
			{"http://nws.api/ilc032/sra", "ilc032", "county"},
		},
		Fails: []SaveZoneFailure{
			{
				SaveZoneResult: SaveZoneResult{"http://nws.api/ilc023/sdr", "ilc023", "county"},
				err:            errors.New("failed to Write"),
			},
		},
		CreatedAt: time.Now().UTC(),
	}, nil
}
