package state

import (
	"context"
	"database/sql"
	"time"

	"github.com/cicconee/weather-app/internal/geometry"
)

type ZoneData struct {
	URI           string
	Code          string
	Type          string
	Name          string
	EffectiveDate time.Time
	State         string
}

func (z *ZoneData) SaveZoneFailure(err error) SaveZoneFailure {
	return SaveZoneFailure{
		URI:  z.URI,
		Code: z.Code,
		Type: z.Type,
		err:  err,
	}
}

type Zone struct {
	Geometry geometry.MultiPolygon
	ZoneData
}

func (z *Zone) ToEntity() ZoneEntity {
	return ZoneEntity{
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		ZoneData:  z.ZoneData,
	}
}

type ZoneEntity struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
	ZoneData
}

func (z *ZoneEntity) Insert(ctx context.Context, db QueryRower) error {
	query := `
		INSERT INTO state_zones(uri, code, type, name, effective_date, state, created_at, updated_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	return db.QueryRowContext(ctx, query,
		z.URI,
		z.Code,
		z.Type,
		z.Name,
		z.EffectiveDate,
		z.State,
		z.CreatedAt,
		z.UpdatedAt,
	).Scan(&z.ID)
}

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
