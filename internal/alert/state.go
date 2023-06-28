package alert

import (
	"context"
	"database/sql"
)

type State string

func (s *State) Scan(scanner Scanner) error {
	return scanner.Scan(s)
}

// State is a collection of states.
type StateCollection []State

// AsStrings converts this state collection
// into a collection of strings.
func (s *StateCollection) AsStrings() []string {
	ss := []string{}

	for _, state := range *s {
		ss = append(ss, string(state))
	}

	return ss
}

// Select reads all the states from the database
// and stores it in this state collection.
func (s *StateCollection) Select(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "SELECT id FROM states")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var state State

		if err := state.Scan(rows); err != nil {
			return err
		}

		*s = append(*s, state)
	}

	return nil
}
