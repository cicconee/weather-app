package state

import (
	"context"
	"database/sql"

	"github.com/cicconee/weather-app/internal/geometry"
)

type PerimeterEntity struct {
	ID     int
	ZoneID int
	Points geometry.PointCollection
}

func (p *PerimeterEntity) Insert(ctx context.Context, db QueryRower) error {
	query := `
		INSERT INTO state_zone_perimeters(sz_id, boundary)
		VALUES($1, $2)
		RETURNING id`

	return db.QueryRowContext(ctx, query,
		p.ZoneID,
		p.Points.String()).Scan(&p.ID)
}

type HoleEntity struct {
	PerimieterID int
	Points       geometry.PointCollection
}

func (h *HoleEntity) Insert(ctx context.Context, db Execer) (sql.Result, error) {
	query := `
		INSERT INTO state_zone_holes(zp_id, boundary)
		VALUES($1, $2)`

	return db.ExecContext(ctx, query, h.PerimieterID, h.Points.String())
}
