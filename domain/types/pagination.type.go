package types

type Pagination struct {
	Page         int
	PageSize     int
	SortBy       string
	SortOrder    string
	FilterParams map[string]string
	Search       *string
}
