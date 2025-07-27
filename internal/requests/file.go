package requests

import "io"

type File struct {
	Name        string
	Reader      io.Reader
	Size        int64
	ContentType string
}
