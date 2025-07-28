package models

type File struct {
	Name string
	URL  string
	Type string
}

type UploadedFile struct {
	Name string
	Path string
	Size int64
}
