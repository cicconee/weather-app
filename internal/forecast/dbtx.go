package forecast

import (
	"context"
	"database/sql"
)

type QueryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}
