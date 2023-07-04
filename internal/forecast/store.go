package forecast

import (
	"context"
	"database/sql"

	"github.com/cicconee/weather-app/internal/geometry"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

func (s *Store) SelectGridpoint(ctx context.Context, point geometry.Point) (GridpointEntity, error) {
	gridpoint := GridpointEntity{}
	return gridpoint, gridpoint.Select(ctx, s.DB, point)
}
