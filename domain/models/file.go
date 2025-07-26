package models

type File struct {
	Name    string
	Path    string
	Type    string
	Content []byte
}

type UploadedFile struct {
	FileName string
	FilePath string
	Size     int64
}
