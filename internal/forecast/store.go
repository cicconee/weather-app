package forecast

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cicconee/weather-app/internal/geometry"
)

// Store is the database storage that can write and read forecast data.
type Store struct {
	// The database connection.
	DB *sql.DB
}

// NewStore creates and returns a Store with the database connection db.
func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

// tx accepts a txFunc and passes it a database transaction. The transaction
// is then commited. If any errors occurs, the transaction will rollback.
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

// SelectGridpoint will read a GridpointEntity from the database where
// the geometric boundary encompasses point. If no rows are found a
// sql.ErrNoRows error is returned with an empty GridpointEntity.
func (s *Store) SelectGridpoint(ctx context.Context, point geometry.Point) (GridpointEntity, error) {
	gridpoint := GridpointEntity{}
	return gridpoint, gridpoint.Select(ctx, s.DB, point)
}

// SelectPeriodCollection reads the PeriodEntity that belong to a gridpoint
// from the database and returns them in a PeriodEntityCollection.
func (s *Store) SelectPeriodCollection(ctx context.Context, gridpointID int) (PeriodEntityCollection, error) {
	periodCollection := PeriodEntityCollection{}
	return periodCollection, periodCollection.Select(ctx, s.DB, gridpointID)
}

// GridpointPeriodsTxParams is the parameters for InsertGridpointPeriodTx.
type GridpointPeriodsTxParams struct {
	Gridpoint *GridpointEntity
	Periods   PeriodEntityCollection
}

// InsertGridpointPeriodsTx writes the GridpointEntity and PeriodEntityCollection
// to the database. The GridpointEntity ID field will be set and all PeriodEntity in
// the PeriodEntityCollection will have the GridpointID field set.
//
// InsertGridpointPeriodsTx is wrapped in a database transaction. If any database
// operations fail, the database will rollback.
func (s *Store) InsertGridpointPeriodsTx(ctx context.Context, p GridpointPeriodsTxParams) error {
	return s.tx(ctx, func(tx *sql.Tx) error {
		if err := p.Gridpoint.Insert(ctx, tx); err != nil {
			return err
		}

		if err := p.Periods.Insert(ctx, tx, p.Gridpoint.ID); err != nil {
			return err
		}

		return nil
	})
}

// UpdateGridpointPeriodTx writes the GridpointEntity and PeriodEntityCollection
// to the database as an update. All the PeriodEntity in the PeriodEntityCollection
// will have the GridpointID field set.
//
// UpdateGridpointPeriodTx is wrapped in a database transaction. If any database
// operation fail, the database will rollback.
func (s *Store) UpdateGridpointPeriodTx(ctx context.Context, p GridpointPeriodsTxParams) error {
	return s.tx(ctx, func(tx *sql.Tx) error {
		if err := p.Gridpoint.Update(ctx, tx); err != nil {
			return err
		}

		if err := p.Periods.Update(ctx, tx, p.Gridpoint.ID); err != nil {
			return err
		}

		return nil
	})
}
