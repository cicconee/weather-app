package forecast

import (
	"context"
	"database/sql"

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

// SelectGridpoint will read a GridpointEntity from the database where
// the geometric boundary encompasses point. If no rows are found a
// sql.ErrNoRows error is returned with an empty GridpointEntity.
func (s *Store) SelectGridpoint(ctx context.Context, point geometry.Point) (GridpointEntity, error) {
	gridpoint := GridpointEntity{}
	return gridpoint, gridpoint.Select(ctx, s.DB, point)
}
