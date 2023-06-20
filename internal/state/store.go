package state

import (
	"context"
	"database/sql"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

func (s *Store) SelectEntity(ctx context.Context, stateID string) (Entity, error) {
	e := Entity{ID: stateID}
	return e, e.Select(ctx, s.DB)
}

func (s *Store) InsertEntity(ctx context.Context, state Entity) (sql.Result, error) {
	return state.Insert(ctx, s.DB)
}
