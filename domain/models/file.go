package models

import "io"

type UploadedFile struct {
	Name string
	Path string
	Size int64
}

// DownloadedFile is a streamable object fetched from storage. Caller must
// close Reader.
type DownloadedFile struct {
	Reader      io.ReadCloser
	ContentType string
	Size        int64
	ETag        string
}
