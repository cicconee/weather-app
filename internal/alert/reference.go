package alert

import (
	"context"
	"database/sql"
)

// Reference is a reference to a alert.
// Reference is used in to defined a alert
// that is outdated and must be deleted.
type Reference string

// Delete deletes the alert reference from
// the database.
//
// Reference must be set before calling this func.
func (r *Reference) Delete(ctx context.Context, db *sql.Tx) (sql.Result, error) {
	return db.ExecContext(ctx, "DELETE FROM alerts WHERE id = $1", r)
}

// Reference is a collection of references of
// alerts. ReferenceCollection is used in
// Resource to defined a collection of alerts
// that are outdated and need to be deleted.
type ReferenceCollection []Reference

// Delete will delete each reference from
// the database.
func (r *ReferenceCollection) Delete(ctx context.Context, db *sql.Tx) error {
	for _, ref := range *r {
		if _, err := ref.Delete(ctx, db); err != nil {
			return err
		}
	}

	return nil
}
