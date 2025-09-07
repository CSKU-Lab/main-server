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
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}
