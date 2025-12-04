package services

import "context"

type FindByService[RES any, REQ any] interface {
	Find(ctx context.Context, req REQ) ([]RES, error)
}
