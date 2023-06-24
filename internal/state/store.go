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

func (s *Store) InsertZoneTx(ctx context.Context, zone Zone) error {
	return s.tx(ctx, func(tx *sql.Tx) error {
		return zone.Insert(ctx, tx)
	})
}
