package requests

import "io"

type File struct {
	Name     string
	Content  io.Reader
	Size     int64
	MimeType string
}
