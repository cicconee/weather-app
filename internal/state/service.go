package state

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type Service struct {
	db *sql.DB
}

func New(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Save(ctx context.Context, stateID string) (SaveResult, error) {
	e := Entity{ID: strings.ToUpper(stateID)}

	err := e.Select(ctx, s.db)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return SaveResult{}, err
	}
	if err == nil {
		return SaveResult{}, errors.New("state already exits")
	}

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
