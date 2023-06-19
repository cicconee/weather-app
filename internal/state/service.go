package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Service struct {
	Store *Store
}

func New(db *sql.DB) *Service {
	return &Service{
		Store: NewStore(db),
	}
}

func (s *Service) Save(ctx context.Context, stateID string) (SaveResult, error) {
	stateID = strings.ToUpper(stateID)

	_, err := s.Store.SelectEntity(ctx, stateID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return SaveResult{}, fmt.Errorf("failed to select state %q: %w", stateID, err)
	}
	if err == nil {
		return SaveResult{}, &Error{
			error:      fmt.Errorf("state %q already saved to database", stateID),
			msg:        fmt.Sprintf("%s already exists", stateID),
			statusCode: http.StatusConflict,
		}
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
