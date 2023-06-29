package alert

import (
	"context"
	"database/sql"
)

// Zone is a geographical location that
// an alert belongs to.
type Zone struct {
	// The identifier of a zone.
	ID int

	// The uri of the zone. This is
	// also a identifier, but not the
	// primary key.
	URI string
}

// Select reads a zone that contains the uri
// from the database and stores it in this zone.
//
// The URI field must be set before calling this
// func.
func (z *Zone) Select(ctx context.Context, db *sql.Tx) error {
	return db.QueryRowContext(ctx, "SELECT id FROM state_zones WHERE uri = $1", z.URI).Scan(&z.ID)
}

// AlertZone is the relationship
// between alerts and zones.
type AlertZone struct {
	// The identifier of the alert.
	AlertID string

	// The identifier of the zone.
	ZoneID int
}

// Insert writes this area zone relationship into
// the database.
//
// AlertID and ZoneID must be set before calling
// this func.
func (a *AlertZone) Insert(ctx context.Context, db *sql.Tx) (sql.Result, error) {
	return db.ExecContext(ctx, "INSERT INTO alert_zones(alert_id, sz_id) VALUES($1, $2)", a.AlertID, a.ZoneID)
}

// LonelyAlert is the relationship
// between a alert and a not yet
// persisted zone.
type LonelyAlert struct {
	// The identifier of the alert.
	AlertID string

	// The uri and identifier of the
	// zone.
	ZoneURI string
}

// Insert writes this lonely area zone relationship
// into the database.
//
// AlertID and ZoneURI must be set before calling
// this func.
func (a *LonelyAlert) Insert(ctx context.Context, db *sql.Tx) (sql.Result, error) {
	return db.ExecContext(ctx, "INSERT INTO lonely_alerts(alert_id, sz_uri) VALUES($1, $2)", a.AlertID, a.ZoneURI)
}
