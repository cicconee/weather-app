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

	w := newWorker(s.Client, s.Pool, s.Store, state.TotalZones)
	defer w.close()

	// Fetch and write each zone to the
	// database.
	zoneResult := w.SaveEach(ctx, zones)

	return SaveResult{
		State:     stateID,
		Writes:    zoneResult.Writes,
		Fails:     zoneResult.Fails,
		CreatedAt: state.CreatedAt,
	}, nil
}

func (s *Service) Sync(ctx context.Context, stateID string) (SaveResult, error) {
	stateID = strings.ToUpper(stateID)

	_, err := s.Store.SelectEntity(ctx, stateID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SaveResult{}, &Error{
				error:      fmt.Errorf("state not found in database (stateID=%q): %w", stateID, err),
				msg:        fmt.Sprintf("%s not found", stateID),
				statusCode: http.StatusNotFound,
			}
		}

		return SaveResult{}, fmt.Errorf("failed to select state in database (stateID=%q): %w", stateID, err)
	}

	updatedZones, err := s.zones(stateID)
	if err != nil {
		return SaveResult{}, fmt.Errorf("failed to get zones (stateID=%q): %w", stateID, err)
	}

	state := &Entity{
		ID:         stateID,
		TotalZones: len(updatedZones),
	}
	if _, err = s.Store.UpdateEntity(ctx, state); err != nil {
		return SaveResult{}, fmt.Errorf("failed to update state %q: %w", stateID, err)
	}

	storedZoneMap := ZoneURIMap{}
	if err := storedZoneMap.Select(ctx, s.Store.DB, stateID); err != nil {
		return SaveResult{}, fmt.Errorf("failed to select zones in database (stateID=%q): %w", stateID, err)
	}

	delta := s.delta(updatedZones, storedZoneMap)

	fmt.Println("INSERT:", delta.Insert)
	fmt.Println("UPDATE:", delta.Update)
	fmt.Println("DELETE:", delta.Delete)

	return SaveResult{}, nil
}

type ZoneDelta struct {
	Insert ZoneCollection
	Update ZoneCollection
	Delete ZoneCollection
}

func NewZoneDelta() *ZoneDelta {
	return &ZoneDelta{
		Insert: ZoneCollection{},
		Update: ZoneCollection{},
		Delete: ZoneCollection{},
	}
}

func (s *Service) delta(updatedZones []Zone, storedZones ZoneURIMap) *ZoneDelta {
	delta := NewZoneDelta()

	for i := range updatedZones {
		updatedZone := updatedZones[i]

		if storedZone, ok := storedZones[updatedZone.URI]; ok {
			if storedZone.EffectiveDate.Before(updatedZone.EffectiveDate) {
				storedZone.CopyUpdateableData(updatedZone)
				delta.Update = append(delta.Update, storedZone)
			}

			delete(storedZones, storedZone.URI)
		} else {
			delta.Insert = append(delta.Insert, updatedZone)
		}
	}

	for uri := range storedZones {
		delta.Delete = append(delta.Delete, storedZones[uri])
	}

	return delta
}

func (s *Service) zones(stateID string) ([]Zone, error) {
	zones, err := s.Client.GetZoneCollection(stateID)
	var statusError *nws.StatusCodeError
	switch {
	case err == nil:
		return zonesFromNWS(zones), nil
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

func zoneFromNWS(z nws.Zone) Zone {
	return Zone{
		URI:           z.URI,
		Code:          z.Code,
		Type:          z.Type,
		Name:          z.Name,
		EffectiveDate: z.EffectiveDate,
		State:         z.State,
		Geometry:      NewGeometry(z.Geometry),
	}
}

func zonesFromNWS(nwsZones []nws.Zone) []Zone {
	zones := []Zone{}
	for i := range nwsZones {
		zones = append(zones, zoneFromNWS(nwsZones[i]))
	}
	return zones
}
