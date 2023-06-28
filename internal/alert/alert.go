package alert

import (
	"context"
	"database/sql"
	"time"

	"github.com/cicconee/weather-app/internal/geometry"
)

// Resource is a alert and all its relationships.
type Resource struct {
	// The alert to be mapped to references
	// and zones.
	Alert *Alert

	// A collection of alert references. Any
	// referenced alert is considered outdated
	// if stored in this collection. If the
	// alert MessageType field is "Update" or
	// "Cancel", this field will be populated.
	// If the alert MessageType field is
	// "Alert" this field will be empty.
	References ReferenceCollection

	// A collection of zones associated with
	// the alert. Any alert with a empty Points
	// field will determine its geometric bounds
	// through these zones.
	Zones []Zone
}

// Alert is a alert for a geographical location.
type Alert struct {
	// The alert identifier.
	ID string

	// A area description. Each area is
	// seperated by a semi colon.
	AreaDesc string

	// The start time of the alert. It is
	// possible to have a nil time.
	OnSet *time.Time

	// The time this alert is expired. If
	// the alert Ends field is nil, Expires
	// will instead be used to determined
	// outdated.
	Expires time.Time

	// The time this alert is determined to
	// be outdated. It is possible to have a
	// nil time.
	Ends *time.Time

	// The alert message type (Alert, Updated,
	// Cancel).
	MessageType string

	// The code denoting the category of the
	// subject event of the alert message
	// (Met, Geo, Safety, Security, Rescue, Fire,
	// Health, Env, Transport, Infra, CBRNE, Other).
	Category string

	// The severity of the alert (Extreme, Severe,
	// Moderate, Minor, Unknown).
	Severity string

	// The chance the alert will occur (Observed,
	// Likely, Possible, Unlikely, Unknown).
	Certainty string

	// The urgency of the alert (Immediate,
	// Expected, Future, Past, Unknown).
	Urgency string

	// The text denoting the type of the subject
	// event of the alert.
	Event string

	// The text headline of the alert. This field
	// may be empty.
	Headline string

	// The text describing the subject event of
	// the alert.
	Description string

	// The text describing the recommended action
	// to be taken by recipients of the alert. This
	// field may be empty.
	Instruction string

	// The code denoting the type of action
	// recommended for the target audience
	// (Shelter, Evacuate, Prepare, Execute,
	// Avoid, Monitor, Assess, AllClear, None).
	Response string

	// The geometric bounds of the alert. This field
	// may be empty.
	Points geometry.Polygon

	// The time the alert was written to the
	// database.
	CreatedAt time.Time
}

func (a *Alert) Scan(scanner Scanner) error {
	return scanner.Scan(
		&a.ID,
		&a.AreaDesc,
		&a.OnSet,
		&a.Expires,
		&a.Ends,
		&a.MessageType,
		&a.Category,
		&a.Severity,
		&a.Certainty,
		&a.Urgency,
		&a.Event,
		&a.Headline,
		&a.Description,
		&a.Instruction,
		&a.Response,
		&a.CreatedAt,
	)
}

// Select reads a alert by id from the database
// and stores it into this alert.
//
// ID must be set before calling this func.
func (a *Alert) Select(ctx context.Context, db *sql.DB) error {
	query := `SELECT id, area_desc, onset, expires, ends, message_type, category, 
			  severity, certainty, urgency, event, headline, description, instruction, 
			  response, created_at FROM alerts WHERE id = $1`

	return a.Scan(db.QueryRowContext(ctx, query, a.ID))
}

// Insert writes this alert into the database.
func (a *Alert) Insert(ctx context.Context, db *sql.Tx) error {
	query := `INSERT INTO alerts(id, area_desc, onset, expires, ends, message_type, category,
			  severity, certainty, urgency, event, headline, description, instruction, response,
			  boundary, created_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 
			  $13, $14, $15, $16, $17)`

	_, err := db.ExecContext(ctx, query,
		a.ID,
		a.AreaDesc,
		a.sqlOnSet(),
		a.Expires,
		a.sqlEnds(),
		a.MessageType,
		a.Category,
		a.Severity,
		a.Certainty,
		a.Urgency,
		a.Event,
		a.Headline,
		a.Description,
		a.Instruction,
		a.Response,
		a.sqlPoints(),
		a.CreatedAt)

	return err
}

func (a *Alert) sqlOnSet() sql.NullTime {
	return a.nullTime(a.OnSet)
}

func (a *Alert) sqlEnds() sql.NullTime {
	return a.nullTime(a.Ends)
}

func (a *Alert) nullTime(t *time.Time) sql.NullTime {
	return sql.NullTime{
		Time:  *t,
		Valid: !t.IsZero() && t != nil,
	}
}

func (a *Alert) sqlPoints() sql.NullString {
	p := a.Points.Permiter()
	return sql.NullString{
		String: p.String(),
		Valid:  p != nil,
	}
}

// AlertCollection is a collection of alerts.
// AlertCollection is used to read and delete
// collections of alerts.
type AlertCollection []Alert

// SelectPointless reads a collection of alerts
// that do not have a defined geometric bounds and
// stores the alerts into this alert collection.
// Each alert associated with a zone that contains
// point will be read.
//
// Alerts with a MessageType of "Cancel" will not
// be read.
func (a *AlertCollection) SelectPointless(ctx context.Context, db *sql.DB, point geometry.Point) error {
	query := `SELECT a.id, a.area_desc, a.onset, a.expires, a.ends, a.message_type, a.category, 
			  a.severity, a.certainty, a.urgency, a.event, a.headline, a.description, a.instruction, 
			  a.response, a.created_at FROM alerts AS a, alert_zones, state_zone_perimeters 
			  WHERE state_zone_perimeters.sz_id = alert_zones.sz_id AND alert_zones.alert_id = a.id
			  AND a.message_type != $1 AND state_zone_perimeters.boundary @> $2`

	rows, err := db.QueryContext(ctx, query, "Cancel", point.String())
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var alert Alert
		if err := alert.Scan(rows); err != nil {
			return err
		}
		*a = append(*a, alert)
	}

	return nil
}

// Select reads a collection of alerts that
// have a defined geometric bounds. Each alert
// geometric bounds that contains point will be
// read.
//
// Alerts with a MessageType of "Cancel" will not
// be read.
func (a *AlertCollection) Select(ctx context.Context, db *sql.DB, point geometry.Point) error {
	query := `SELECT id, area_desc, onset, expires, ends, message_type, category, 
			  severity, certainty, urgency, event, headline, description, instruction, 
			  response, created_at FROM alerts WHERE message_type != $1 AND boundary @> $2`

	rows, err := db.QueryContext(ctx, query, "Cancel", point.String())
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var alert Alert
		if err := alert.Scan(rows); err != nil {
			return err
		}
		*a = append(*a, alert)
	}

	return nil
}

// DeleteEnded will delete all alerts from the
// database that has ended before t.
func (e *AlertCollection) DeleteEnded(ctx context.Context, db *sql.DB, t time.Time) (sql.Result, error) {
	return db.ExecContext(ctx, "DELETE FROM alerts WHERE ends < $1", t)
}

// DeleteExpired will delete all alerts from
// the database that has expired before t.
func (e *AlertCollection) DeleteExpired(ctx context.Context, db *sql.DB, t time.Time) (sql.Result, error) {
	return db.ExecContext(ctx, "DELETE FROM alerts WHERE ends IS NULL AND expires < $1", t)
}
