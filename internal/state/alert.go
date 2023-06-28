package state

import "context"

// AlertZone is a alert that falls in the
// boundary of a zone and the zone is persisted
// to the database.
type AlertZone struct {
	AlertID string
	ZoneID  int
}

// Insert will write this alert zone to the database.
//
// AlertID and ZoneID must be set before calling
// Insert.
func (a *AlertZone) Insert(ctx context.Context, db Execer) error {
	query := "INSERT INTO alert_zones(alert_id, sz_id) VALUES($1, $2)"
	_, err := db.ExecContext(ctx, query, a.AlertID, a.ZoneID)
	return err
}

// LonelyAlert is a alert that fall in the
// boundary of a zone, but the zone is not
// yet persisted in the database. Due to alerts
// crossing state boundaries, it is possible
// that alerts associated with a supported
// state also includes another state that is
// not yet supported. By keeping a lonely alert
// log, as soon as a new state is supported all
// lonely alerts can immediately be mapped to
// the appropriate zones.
type LonelyAlert struct {
	AlertID string
	ZoneURI string
}

func (a *LonelyAlert) scan(scanner func(...any) error) error {
	return scanner(&a.AlertID, &a.ZoneURI)
}

// Delete will delete this lonely alert from the
// database.
//
// AlertID and ZoneURI must be set before calling
// DeleteLonely.
func (a *LonelyAlert) Delete(ctx context.Context, db Execer) error {
	query := "DELETE FROM lonely_alerts WHERE alert_id = $1 AND sz_uri = $2"
	_, err := db.ExecContext(ctx, query, a.AlertID, a.ZoneURI)
	return err
}

// LonelyAlertCollection is a collection of
// lonely alerts.
//
// LonelyAlertCollection is used to select
// a collection of lonely alerts from the database.
type LonelyAlertCollection []LonelyAlert

// Select will select all the lonely alerts
// for a zone uri (zoneURI) and store them in
// this LonelyAlertCollection.
//
// Each lonely alert in the collection will
// have its AlertID and ZoneURI set.
func (a *LonelyAlertCollection) Select(ctx context.Context, db Queryer, zoneURI string) error {
	query := `SELECT alert_id, sz_uri FROM lonely_alerts WHERE sz_uri = $1`

	rows, err := db.QueryContext(ctx, query, zoneURI)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		alert := LonelyAlert{}
		if alert.scan(rows.Scan); err != nil {
			return err
		}
		*a = append(*a, alert)
	}

	return nil
}
