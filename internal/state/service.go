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
	"github.com/cicconee/weather-app/internal/pool"
)

type Service struct {
	Client *nws.Client
	Store  *Store
	Pool   *pool.Pool
}

func New(c *nws.Client, db *sql.DB, p *pool.Pool) *Service {
	return &Service{
		Client: c,
		Store:  NewStore(db),
		Pool:   p,
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

	zones, err := s.zones(stateID)
	if err != nil {
		return SaveResult{}, fmt.Errorf("failed to get zones for %q: %w", stateID, err)
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

	w := newWorker(s.Client, s.Pool, state.TotalZones)
	defer w.close()

	// Queue each zone in the worker to get
	// the data concurrently.
	w.FetchEach(ctx, zonesFromNWS(zones))

	// Process each fetched zone and record if
	// the zone was successfully writen or failed.
	writes, fails := s.process(ctx, state.TotalZones, w)

	return SaveResult{
		State:     stateID,
		Writes:    writes,
		Fails:     fails,
		CreatedAt: state.CreatedAt,
	}, nil
}

func (s *Service) zones(stateID string) ([]nws.Zone, error) {
	zones, err := s.Client.GetZoneCollection(stateID)
	var statusError *nws.StatusCodeError
	switch {
	case err == nil:
		return zones, nil
	case errors.As(err, &statusError):
		if statusError.StatusCode == 400 {
			return nil, &Error{
				error:      fmt.Errorf("unsupported state: %w", err),
				msg:        fmt.Sprintf("%s is not a valid state", stateID),
				statusCode: http.StatusNotFound,
			}
		}

		return nil, fmt.Errorf("unexpected status code: %w", err)
	default:
		return nil, err
	}
}

func (s *Service) process(ctx context.Context, n int, w *worker) ([]SaveZoneResult, []SaveZoneFailure) {
	// Initialize slices that will hold
	// write results for zones.
	writes := []SaveZoneResult{}
	fails := []SaveZoneFailure{}

	// Write each successful zone fetch
	// to the database. If any errors
	// record it in the fails slice.
	for i := 0; i < n; i++ {
		select {
		case zone := <-w.dataCh:
			saveZoneResult := SaveZoneResult{
				URI:  zone.URI,
				Code: zone.Code,
				Type: zone.Type,
			}

			err := s.Store.InsertZoneTx(ctx, zone)
			if err != nil {
				fails = append(fails, SaveZoneFailure{
					SaveZoneResult: saveZoneResult,
					err:            err,
				})
			} else {
				writes = append(writes, saveZoneResult)
			}
		case fail := <-w.failCh:
			fails = append(fails, fail)
		}
	}

	return writes, fails
}

func zoneFromNWS(z nws.Zone) Zone {
	return Zone{
		Geometry: z.Geometry,
		ZoneData: ZoneData{
			URI:           z.URI,
			Code:          z.Code,
			Type:          z.Type,
			Name:          z.Name,
			EffectiveDate: z.EffectiveDate,
			State:         z.State,
		},
	}
}

func zonesFromNWS(nwsZones []nws.Zone) []Zone {
	zones := []Zone{}
	for i := range nwsZones {
		zones = append(zones, zoneFromNWS(nwsZones[i]))
	}
	return zones
}
