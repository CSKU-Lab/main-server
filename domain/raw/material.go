package raw

import "time"

type Material struct {
	ID         string    `db:"id"`
	Name       string    `db:"name"`
	Type       string    `db:"type"`
	Visibility string    `db:"visibility"`
	CreatedAt  time.Time `db:"created_at"`
	CreatedBy  string    `db:"created_by"`
}
