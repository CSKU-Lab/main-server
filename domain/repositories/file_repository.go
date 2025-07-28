package repositories

import (
	"context"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
)

type FileRepository interface {
	UploadFile(ctx context.Context, path string, file *requests.File) (*models.UploadedFile, error)
	UploadMultipleFiles(ctx context.Context, path string, files []requests.File) ([]models.UploadedFile, error)
	DeleteFile(ctx context.Context, ID string) error
	DeleteManyFiles(ctx context.Context, paths []string) error
}
