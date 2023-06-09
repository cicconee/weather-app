package state

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

func (s *Store) tx(ctx context.Context, txFunc func(*sql.Tx) error) error {
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	if err := txFunc(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("err: %w, rbErr: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

func (s *Store) SelectEntity(ctx context.Context, stateID string) (Entity, error) {
	e := Entity{ID: stateID}
	return e, e.Select(ctx, s.DB)
}

func (s *Store) InsertEntity(ctx context.Context, state Entity) (sql.Result, error) {
	return state.Insert(ctx, s.DB)
}

// UpdateEntity writes state to the database
// as an update. The state UpdatedAt field will be
// set to the current time in UTC format before
// writing to the database.
//
// If the state UpdatedAt field is set it will be
// overwritten.
func (s *Store) UpdateEntity(ctx context.Context, state *Entity) (sql.Result, error) {
	state.UpdatedAt = time.Now().UTC()
	return state.Update(ctx, s.DB)
}

// SelectZonesWhereState selects all the zones
// for a given state (stateID) as a ZoneURIMap.
func (s *Store) SelectZonesWhereState(ctx context.Context, stateID string) (ZoneURIMap, error) {
	storedZoneMap := ZoneURIMap{}
	return storedZoneMap, storedZoneMap.Select(ctx, s.DB, stateID)
}

// InsertZoneTx writes zone to the database.
// The zone ID, CreatedAt, and UpdatedAt field
// will be set. If these are set before calling
// the func, they will be ignored and overwritten.
//
// When a zone is inserted, any lonely alerts
// associated with zone will be deleted from the
// database and transformed into an alert zone.
// Lonely alerts are designed to hold the place of
// an alert zone before the zone is created. Since
// the zone is being created, the alerts are no
// longer lonely.
//
// InsertZoneTx is wrapped in a database transaction.
// If any operations fail the database will roll back.
func (s *Store) InsertZoneTx(ctx context.Context, zone *Zone) error {
	return s.tx(ctx, func(tx *sql.Tx) error {
		if err := zone.Insert(ctx, tx); err != nil {
			return err
		}

		collection := LonelyAlertCollection{}
		if err := collection.Select(ctx, tx, zone.URI); err != nil {
			return err
		}

		for _, lonely := range collection {
			alert := AlertZone{AlertID: lonely.AlertID, ZoneID: zone.ID}
			if err := alert.Insert(ctx, tx); err != nil {
				return fmt.Errorf("failed to insert alert zone (AlertID=%s, ZoneID=%d): %w", alert.AlertID, alert.ZoneID, err)
			}

			if err := lonely.Delete(ctx, tx); err != nil {
				return fmt.Errorf("failed to delete lonely alert: %w", err)
			}
		}

		return nil
	})
}

// UpdateZoneTx writes zone to the database as
// a update. The current Geometry stored in the
// database for zone will be deleted. The Geometry
// stored in the Geometry field of zone will then
// be inserted to the database.
//
// UpdateZoneTx is wrapped in a database transaction.
// If any operations fail the database will roll back.
//
// If the zone UpdatedAt field is set it will be
// overwritten.
func (s *Store) UpdateZoneTx(ctx context.Context, zone *Zone) error {
	return s.tx(ctx, func(tx *sql.Tx) error {
		zone.UpdatedAt = time.Now().UTC()
		return zone.Update(ctx, tx)
	})
}

// DeleteZone deletes the zone with the provided
// ID (zoneID).
func (s *Store) DeleteZone(ctx context.Context, zoneID int) error {
	zone := Zone{ID: zoneID}
	_, err := zone.Delete(ctx, s.DB)
	return err
}
