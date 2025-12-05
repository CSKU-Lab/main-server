package sanitize

import (
	"net/http"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
)

func SortOrder(sortOrder string) (string, error) {
	if sortOrder != "asc" && sortOrder != "desc" {
		return "", cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Invalid sort order",
		})
	}
	return sortOrder, nil
}

func SortBy(sortBy string, fields map[string]bool) (string, error) {
	if !fields[sortBy] {
		return "", cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "The field that you want to sort by is not exist",
		})
	}

	return sortBy, nil
}

type Filter struct {
	Field    string
	Operator string
	Value    string
}

func FindBy(allowedFields map[string]bool, field string) error {
	_, ok := allowedFields[field]
	if !allowedFields[field] || !ok {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Invalid field: " + field,
		})
	}

	return nil
}

func Filters(filterParams map[string]string, allowedFields map[string]bool) ([]Filter, error) {
	allowedOperators := map[string]bool{
		"is":           true,
		"is_not":       true,
		"contains":     true,
		"not_contains": true,
		"is_empty":     true,
		"is_not_empty": true,
		"gt":           true,
		"gte":          true,
		"lt":           true,
		"lte":          true,
	}

	var filters []Filter

	for key, value := range filterParams {
		if value == "" {
			continue
		}

		parts := strings.Split(key, "__")
		if len(parts) != 2 {
			continue
		}

		field := parts[0]
		operator := parts[1]

		if !allowedFields[field] {
			return nil, cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid filter field: " + field,
			})
		}

		if !allowedOperators[operator] {
			return nil, cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid filter operator: " + operator,
			})
		}

		filters = append(filters, Filter{
			Field:    field,
			Operator: operator,
			Value:    value,
		})
	}

	return filters, nil
}
