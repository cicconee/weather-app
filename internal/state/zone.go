package state

import (
	"context"
	"database/sql"
	"time"
)

type Zone struct {
	ID            int
	URI           string
	Code          string
	Type          string
	Name          string
	EffectiveDate time.Time
	State         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Geometry      Geometry
}

func (z *Zone) CopyUpdateableData(c Zone) {
	z.URI = c.URI
	z.Code = c.Code
	z.Type = c.Type
	z.Name = c.Name
	z.EffectiveDate = c.EffectiveDate
	z.State = c.State
	z.Geometry = c.Geometry
}

func (z *Zone) SaveZoneFailure(err error) SaveZoneFailure {
	return SaveZoneFailure{
		URI: z.URI,
		err: err,
	}
}

func (z *Zone) Insert(ctx context.Context, db QueryRower) error {
	query := `
		INSERT INTO state_zones(uri, code, type, name, effective_date, state, created_at, updated_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	z.CreatedAt = time.Now().UTC()
	z.UpdatedAt = time.Now().UTC()

	if err := db.QueryRowContext(ctx, query,
		z.URI,
		z.Code,
		z.Type,
		z.Name,
		z.EffectiveDate,
		z.State,
		z.CreatedAt,
		z.UpdatedAt,
	).Scan(&z.ID); err != nil {
		return err
	}

	for _, perimeter := range z.Geometry {
		perimeter.ZoneID = z.ID

		if err := perimeter.Insert(ctx, db); err != nil {
			return err
		}
	}

	return nil
}

func (z *Zone) scan(scanFunc func(...any) error) error {
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

type ZoneCollection []Zone

type ZoneURIMap map[string]Zone

func (z ZoneURIMap) Select(ctx context.Context, db *sql.DB, state string) error {
	query := `
		SELECT id, uri, code, type, name, effective_date, state, created_at, updated_at
		FROM state_zones
		WHERE state = $1`

	rows, err := db.QueryContext(ctx, query, state)
	if err != nil {
		return err
	}

	for rows.Next() {
		var e Zone
		if err := e.scan(rows.Scan); err != nil {
			return err
		}

		z[e.URI] = e
	}

	return nil
}
