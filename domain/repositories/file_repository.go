package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type FileRepository interface {
	UploadFile(ctx context.Context, path string, file *requests.File) (*models.UploadedFile, error)
	UploadMultipleFiles(ctx context.Context, path string, files []requests.File) ([]models.UploadedFile, error)
	DeleteFile(ctx context.Context, path string) error
	DeleteManyFiles(ctx context.Context, paths []string) error
}
