package state

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Entity struct {
	ID           string
	TotalZones   int
	WrittenZones int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (e *Entity) Select(ctx context.Context, db *sql.DB) error {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s",
		"id, total_zones, (SELECT COUNT(*) FROM state_zones WHERE state = $1), created_at, updated_at",
		"states",
		"id = $1")

	return db.QueryRowContext(ctx, query, e.ID).Scan(
		&e.ID,
		&e.TotalZones,
		&e.WrittenZones,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
}

func (e *Entity) Insert(ctx context.Context, db *sql.DB) (sql.Result, error) {
	query := "INSERT INTO states(id, total_zones, created_at, updated_at) VALUES($1, $2, $3, $4)"

	return db.ExecContext(ctx, query,
		e.ID,
		e.TotalZones,
		e.CreatedAt,
		e.UpdatedAt)
}
