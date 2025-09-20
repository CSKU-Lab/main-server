package sqlx

import (
	"fmt"
	"strings"

	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

func buildFilterWhereClause(filters []sanitize.Filter, startingArgIndex int) (string, []any) {
	if len(filters) == 0 {
		return "", nil
	}

	var conditions []string
	var args []any
	argIndex := startingArgIndex

	for _, filter := range filters {
		switch filter.Operator {
		case "is":
			conditions = append(conditions, fmt.Sprintf("%s = $%d", filter.Field, argIndex))
			args = append(args, filter.Value)
			argIndex++

		case "is_not":
			conditions = append(conditions, fmt.Sprintf("%s != $%d", filter.Field, argIndex))
			args = append(args, filter.Value)
			argIndex++

		case "contains":
			conditions = append(conditions, fmt.Sprintf("%s ILIKE $%d", filter.Field, argIndex))
			args = append(args, "%"+filter.Value+"%")
			argIndex++

		case "not_contains":
			conditions = append(conditions, fmt.Sprintf("%s NOT ILIKE $%d", filter.Field, argIndex))
			args = append(args, "%"+filter.Value+"%")
			argIndex++

		case "is_empty":
			conditions = append(conditions, fmt.Sprintf("(%s IS NULL OR %s = '')", filter.Field, filter.Field))

		case "is_not_empty":
			conditions = append(conditions, fmt.Sprintf("(%s IS NOT NULL AND %s != '')", filter.Field, filter.Field))

		case "gt":
			conditions = append(conditions, fmt.Sprintf("%s > $%d", filter.Field, argIndex))
			args = append(args, filter.Value)
			argIndex++

		case "gte":
			conditions = append(conditions, fmt.Sprintf("%s >= $%d", filter.Field, argIndex))
			args = append(args, filter.Value)
			argIndex++

		case "lt":
			conditions = append(conditions, fmt.Sprintf("%s < $%d", filter.Field, argIndex))
			args = append(args, filter.Value)
			argIndex++

		case "lte":
			conditions = append(conditions, fmt.Sprintf("%s <= $%d", filter.Field, argIndex))
			args = append(args, filter.Value)
			argIndex++
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return " AND " + strings.Join(conditions, " AND "), args
}