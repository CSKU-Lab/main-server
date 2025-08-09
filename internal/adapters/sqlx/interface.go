package sqlx

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type instance interface {
	sqlx.Ext
	sqlx.ExtContext
	sqlx.PreparerContext
	sqlx.Preparer
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
