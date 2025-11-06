package sqlx

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type instance interface {
	sqlx.Ext
	sqlx.ExtContext
	sqlx.PreparerContext
	sqlx.Preparer
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}
