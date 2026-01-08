package services

import (
	"context"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/converter"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type MaterialAssetService interface {
	UploadFile(ctx context.Context, materialID string, file *requests.File) (string, error)
}

type materialAssetService struct {
	fileRepo repositories.FileRepository
	configs  *configs.Config
}

func NewMaterialAssetService(configs *configs.Config, fileRepo repositories.FileRepository) MaterialAssetService {
	return &materialAssetService{
		configs:  configs,
		fileRepo: fileRepo,
	}
}

func (s *materialAssetService) UploadFile(ctx context.Context, materialID string, file *requests.File) (string, error) {
	uploadedFile, err := s.fileRepo.UploadFile(ctx, "materials/"+materialID, file)
	if err != nil {
		return "", err
	}

	fileURL := converter.ToS3Path(s.configs, uploadedFile.Path)

	return fileURL, nil
}
