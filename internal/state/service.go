package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cicconee/weather-app/internal/nws"
)

type Service struct {
	Client *nws.Client
	Store  *Store
}

func New(c *nws.Client, db *sql.DB) *Service {
	return &Service{
		Client: c,
		Store:  NewStore(db),
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

	zones, err := s.Client.GetZoneCollection(stateID)
	if err != nil {
		return SaveResult{}, err
	}

	state := Entity{
		ID:         stateID,
		TotalZones: len(zones),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if _, err = s.Store.InsertEntity(ctx, state); err != nil {
		return SaveResult{}, fmt.Errorf("failed to insert state %q: %w", stateID, err)
	}

	return SaveResult{
		State:     stateID,
		Writes:    []SaveZoneResult{},
		Fails:     []SaveZoneFailure{},
		CreatedAt: state.CreatedAt,
	}, nil
}
