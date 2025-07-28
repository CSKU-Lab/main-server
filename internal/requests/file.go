package requests

import (
	"io"
	"strings"
)

type File struct {
	Name        string
	Reader      io.Reader
	Size        int64
	ContentType string
}

func (f *File) Extension() string {
	if len(f.Name) == 0 {
		return ""
	}
	parts := strings.Split(f.Name, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}
