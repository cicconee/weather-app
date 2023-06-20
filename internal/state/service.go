package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
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
		log.Println(err)
		return SaveResult{}, err
	}

	fmt.Println(zones)

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
