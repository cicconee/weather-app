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

func (z *ZoneEntity) scan(scanFunc func(...any) error) error {
	return scanFunc(
		&z.ID,
		&z.URI,
		&z.Code,
		&z.Type,
		&z.Name,
		&z.EffectiveDate,
		&z.State,
		&z.CreatedAt,
		&z.UpdatedAt,
	)
}

type ZoneEntityCollection []ZoneEntity

type ZoneEntityURIMap map[string]ZoneEntity

func (z ZoneEntityURIMap) Select(ctx context.Context, db *sql.DB, state string) error {
	query := `
		SELECT id, uri, code, type, name, effective_date, state, created_at, updated_at
		FROM state_zones
		WHERE state = $1`

	rows, err := db.QueryContext(ctx, query, state)
	if err != nil {
		return err
	}

	for rows.Next() {
		var e ZoneEntity
		if err := e.scan(rows.Scan); err != nil {
			return err
		}

		z[e.URI] = e
	}

	return nil
}
