package sqlx

import (
	"context"
	"fmt"
	"time"

	contextkeys "github.com/CSKU-Lab/main-server/context_keys"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/lib/pq"
)

type sectionLogRepository struct {
	db instance
}

type sectionLog struct {
	ID               string         `db:"id"`
	Timestamp        time.Time      `db:"timestamp"`
	Action           string         `db:"action"`
	IPAddress        string         `db:"ip_address"`
	UserID           string         `db:"user.id"`
	Username         string         `db:"user.username"`
	UserDisplayName  string         `db:"user.display_name"`
	UserProfileImage *string        `db:"user.profile_image"`
	UserRoles        pq.StringArray `db:"user.roles"`
}

func NewSectionLogRepository(db instance) repositories.SectionLogRepository {
	return &sectionLogRepository{db: db}
}

func (s *sectionLogRepository) Create(ctx context.Context, id string, sectionID string, action string) error {
	query := "INSERT INTO section_logs (id, section_id, user_id, action, ip_address) VALUES ($1,$2,$3,$4,$5)"

	userID := ctx.Value(contextkeys.UserKey).(contextkeys.User).ID
	ipAddress := ctx.Value(contextkeys.UserKey).(contextkeys.User).IP_Address

	_, err := s.db.ExecContext(ctx, query, id, sectionID, userID, action, ipAddress)
	if err != nil {
		return err
	}

	return nil
}

func (s *sectionLogRepository) GetPaginationBySectionID(ctx context.Context, sectionID string, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter) ([]models.SectionLog, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 2)
	baseQuery := `
	SELECT
		sl.id,
		sl.timestamp,
		sl.action,
		u.id AS "user.id",
		u.username AS "user.username",
		u.display_name AS "user.display_name",
		u.profile_image AS "user.profile_image",
		u.roles AS "user.roles",
		sl.ip_address
	FROM
		section_logs sl
	JOIN
		users u ON sl.user_id = u.id
	WHERE
		sl.section_id = $1
	`
	query := fmt.Sprintf(`%s %s 
		ORDER BY %s %s
		OFFSET $%d 
		LIMIT $%d
		`, baseQuery, filterWhereClause, sortBy, sortOrder, len(filterArgs)+2, len(filterArgs)+3)

	args := []any{sectionID}
	args = append(args, (page-1)*limit)
	args = append(args, limit)
	args = append(args, filterArgs...)

	var sectionLogs []sectionLog
	err := s.db.SelectContext(ctx, &sectionLogs, query, args...)
	if err != nil {
		return nil, err
	}

	result := make([]models.SectionLog, 0, len(sectionLogs))
	for _, log := range sectionLogs {
		sectionLog := models.SectionLog{
			ID:        log.ID,
			Timestamp: log.Timestamp,
			Action:    log.Action,
			User: models.SectionLogUser{
				Username:     log.Username,
				DisplayName:  log.UserDisplayName,
				ProfileImage: log.UserProfileImage,
				Roles:        log.UserRoles,
			},
			IPAddress: log.IPAddress,
		}
		result = append(result, sectionLog)
	}

	return result, nil
}

func (s *sectionLogRepository) CountBySectionID(ctx context.Context, sectionID string, search string, filters []sanitize.Filter) (int, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 2)
	baseQuery := `
	SELECT
		COUNT(*)
	FROM
		section_logs sl
	JOIN
		users u ON sl.user_id = u.id
	WHERE
		sl.section_id = $1
	`
	query := fmt.Sprintf(`%s %s`, baseQuery, filterWhereClause)

	args := []any{sectionID}
	args = append(args, filterArgs...)

	var count int
	err := s.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}
