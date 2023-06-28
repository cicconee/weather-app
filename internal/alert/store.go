package alert

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cicconee/weather-app/internal/geometry"
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

// SelectAlert reads an alert by id from the
// database.
func (s *Store) SelectAlert(ctx context.Context, id string) (Alert, error) {
	alert := Alert{ID: id}
	return alert, alert.Select(ctx, s.DB)
}

// SelectAlertsContains reads a collection of alerts
// where the point resides inside the boundary of the
// alerts.
//
// The boundary of an alert is determined by either
// the alert having an explicit boundary, or the
// boundary of the zones related to the alert.
func (s *Store) SelectAlertsContains(ctx context.Context, point geometry.Point) (AlertCollection, error) {
	collection := AlertCollection{}

	// Get all the alerts where the geometric bounds
	// are determined through the mapping to zones.
	if err := collection.Select(ctx, s.DB, point); err != nil {
		return AlertCollection{}, err
	}

	// Get all the alerts that have a specified
	// geometric bounds.
	if err := collection.SelectPointless(ctx, s.DB, point); err != nil {
		return AlertCollection{}, err
	}

	return collection, nil
}

// SelectStates reads a collection of states
// from the database. All states in the database
// will reside in this collection.
func (s *Store) SelectStates(ctx context.Context) (StateCollection, error) {
	var collection StateCollection
	return collection, collection.Select(ctx, s.DB)
}

// InsertAlertTx writes an alert resource to the
// database. All alerts persisted to the database
// that are referenced by the resource will be
// deleted. All relationships between the alert
// and zones will be written to the database.
//
// The alert CreatedAt field will be set.
//
// InsertAlertTx is wrapped in a database transaction.
// If any operations fail the database will roll back.
func (s *Store) InsertAlertTx(ctx context.Context, r Resource) error {
	return s.tx(ctx, func(tx *sql.Tx) error {
		r.Alert.CreatedAt = time.Now().UTC()
		if err := r.Alert.Insert(ctx, tx); err != nil {
			return err
		}

		if err := r.References.Delete(ctx, tx); err != nil {
			return err
		}

		for _, z := range r.Zones {
			if err := z.Select(ctx, tx); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}

			var insertErr error
			switch z.ID {
			case 0: // The zone is not stored in the database.
				lonely := LonelyAlert{AlertID: r.Alert.ID, ZoneURI: z.URI}
				_, insertErr = lonely.Insert(ctx, tx)
			default: // The zone is stored in the database.
				alertZone := AlertZone{AlertID: r.Alert.ID, ZoneID: z.ID}
				_, insertErr = alertZone.Insert(ctx, tx)
			}
			if insertErr != nil {
				return insertErr
			}
		}

		return nil
	})
}

// DeleteEndedAlerts will delete all alerts where
// the end time is before t.
func (s *Store) DeleteEndedAlerts(ctx context.Context, t time.Time) (int64, error) {
	collection := AlertCollection{}

	res, err := collection.DeleteEnded(ctx, s.DB, t)
	if err != nil {
		return 0, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return n, nil
}

// DeleteExpiredAlerts will delete all alerts where
// the expire time is before t
func (s *Store) DeleteExpiredAlerts(ctx context.Context, t time.Time) (int64, error) {
	collection := AlertCollection{}
	res, err := collection.DeleteExpired(ctx, s.DB, t)
	if err != nil {
		return 0, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return n, nil
}
