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

type SyncResult struct {
	State   string
	Inserts []Zone
	Updates []Zone
	Deletes []Zone
	Fails   []SyncZoneFailure
}

type SyncZoneFailure struct {
	URI string
	Op  string
	err error
}

func (s *Service) Sync(ctx context.Context, stateID string) (SyncResult, error) {
	stateID = strings.ToUpper(stateID)

	// Selext state from database to make
	// sure it exists.
	state, err := s.Store.SelectEntity(ctx, stateID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SyncResult{}, &Error{
				error:      fmt.Errorf("state not found in database (stateID=%q): %w", stateID, err),
				msg:        fmt.Sprintf("%s not found", stateID),
				statusCode: http.StatusNotFound,
			}
		}

		return SyncResult{}, fmt.Errorf("failed to select state in database (stateID=%q): %w", stateID, err)
	}

	// Get the up to date data for zones.
	// At this point every Zone in updatedZones
	// has an unset Geometry.
	updatedZones, err := s.zones(stateID)
	if err != nil {
		return SyncResult{}, fmt.Errorf("failed to get zones (stateID=%q): %w", stateID, err)
	}

	// Write the state updates to the
	// database.
	state.TotalZones = len(updatedZones)
	if _, err = s.Store.UpdateEntity(ctx, &state); err != nil {
		return SyncResult{}, fmt.Errorf("failed to update state (state.ID=%q): %w", state.ID, err)
	}

	// Get the current zone data from
	// the database to compare to the
	// up to date zone data. This will
	// be used to determine the zone
	// delta (insert, update, delete).
	storedZoneMap, err := s.Store.SelectZonesWhereState(ctx, stateID)
	if err != nil {
		return SyncResult{}, fmt.Errorf("failed to select zones in database (stateID=%q): %w", stateID, err)
	}

	return s.writeDelta(ctx, writeDeltaParams{
		stateID:      stateID,
		updatedZones: updatedZones,
		storedZones:  storedZoneMap,
	}), nil
}

type writeDeltaParams struct {
	stateID      string
	updatedZones []Zone
	storedZones  ZoneURIMap
}

// writeDelta compares the collection of up to date zones
// (updatedZones) to the stored collection of zones (storedZones).
// By comparing these two collections a ZoneDelta is formed
// that specificies what zones need to be inserted, updated,
// or deleted. These changes are then executed to bring the
// database up to date. For any zones needed to be inserted
// or updated, additional network calls are made concurrently.
//
// Any errors that occur while fetching the data or
// persisting the data will be recorded as a SyncZoneFailure
// and stored in the SyncResult.Fails field.
func (s *Service) writeDelta(ctx context.Context, p writeDeltaParams) SyncResult {
	delta := s.delta(p.updatedZones, p.storedZones)

	fetcher := NewFetcher(s.Client, s.Pool, s.Store, delta.TotalInsertUpdates())
	defer fetcher.close()

	// For every zone that needs to be
	// inserted or updated in the database,
	// get the up to date Geometry.
	fetchResult := fetcher.FetchEach(ctx, delta.InsertUpdate())

	result := SyncResult{
		State:   p.stateID,
		Inserts: []Zone{},
		Updates: []Zone{},
		Deletes: []Zone{},
		Fails:   []SyncZoneFailure{},
	}

	// Record any errors while fetching the
	// geometric data.
	for uri, err := range fetchResult.Fails {
		result.Fails = append(result.Fails, SyncZoneFailure{
			URI: uri,
			Op:  "fetch",
			err: err,
		})
	}

	// Insert all the new zones.
	for _, zone := range delta.Insert {
		if z, ok := fetchResult.Zones[zone.URI]; ok {
			if err := s.Store.InsertZoneTx(ctx, &z); err != nil {
				result.Fails = append(result.Fails, SyncZoneFailure{
					URI: z.URI,
					Op:  "insert",
					err: err,
				})
			} else {
				result.Inserts = append(result.Inserts, z)
			}
		}
	}

	// Updated all the expired zones.
	for _, zone := range delta.Update {
		if z, ok := fetchResult.Zones[zone.URI]; ok {
			if err := s.Store.UpdateZoneTx(ctx, &z); err != nil {
				result.Fails = append(result.Fails, SyncZoneFailure{
					URI: z.URI,
					Op:  "update",
					err: err,
				})
			} else {
				result.Updates = append(result.Updates, z)
			}
		}
	}

	// Delete all the old zones.
	for i, zone := range delta.Delete {
		if err := s.Store.DeleteZone(ctx, zone.ID); err != nil {
			result.Fails = append(result.Fails, SyncZoneFailure{
				URI: zone.URI,
				Op:  "delete",
				err: err,
			})
		} else {
			result.Deletes = append(result.Deletes, delta.Delete[i])
		}
	}

	return result
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
