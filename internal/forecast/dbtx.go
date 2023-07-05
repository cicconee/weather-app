package forecast

import (
	"context"
	"database/sql"
)

type Queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

type QueryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}
