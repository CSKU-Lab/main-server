package services

import (
	"net/http"
	"slices"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
)

func sanitizeSortOrder(sortOrder string) (string, error) {
	if sortOrder != "asc" && sortOrder != "desc" {
		return "", cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Invalid sort order",
		})
	}
	return sortOrder, nil
}

func sanitizeSortBy(sortBy string, fields []string) (string, error) {
	if !slices.Contains(fields, sortBy) {
		return "", cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "The field that you want to sort by is not exist",
		})
	}

	return sortBy, nil
}
