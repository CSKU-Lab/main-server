package contextkeys

type userContextKey string

type User struct {
	Username    string
	DisplayName string
	ID          string
	IP_Address  string
}

const (
	UserKey userContextKey = "user"
)
