package sqlx

import (
	"context"
	"fmt"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
)

type sqlxTagRepository struct {
	db *sqlx.DB
}

func NewTagRepository(db *sqlx.DB) repositories.TagRepository {
	return &sqlxTagRepository{db: db}
}

func (r *sqlxTagRepository) GetPagination(ctx context.Context, page int, limit int, search string) ([]models.Tag, error) {
	query := fmt.Sprintf(`SELECT id, name FROM tags
		WHERE name ILIKE $1
		ORDER BY name ASC
		OFFSET $2 LIMIT $3`)

	type tagRow struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}

	var rows []tagRow
	if err := r.db.SelectContext(ctx, &rows, query, "%"+search+"%", (page-1)*limit, limit); err != nil {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Failed to get tags",
		})
	}

	tags := make([]models.Tag, len(rows))
	for i, row := range rows {
		tags[i] = models.Tag{ID: row.ID, Name: row.Name}
	}

	return tags, nil
}

func (r *sqlxTagRepository) Count(ctx context.Context, search string) (int, error) {
	query := `SELECT COUNT(*) FROM tags WHERE name ILIKE $1`

	var count int
	if err := r.db.GetContext(ctx, &count, query, "%"+search+"%"); err != nil {
		return 0, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Failed to count tags",
		})
	}

	return count, nil
}
