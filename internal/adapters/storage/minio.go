package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minIO struct {
	bucket string
	client *minio.Client
}

func NewMinIOStorage(ctx context.Context, config *configs.Config) repositories.FileRepository {
	client, err := minio.New(config.MinIO_Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.MinIO_AccessKeyID, config.MinIO_SecretAccessKey, ""),
		Secure: config.MinIO_UseSSL,
	})
	if err != nil {
		log.Fatalf("❌ Error creating MinIO client: %v", err)
	}

	err = client.MakeBucket(ctx, config.MinIO_Bucket, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := client.BucketExists(ctx, config.MinIO_Bucket)
		if errBucketExists != nil {
			log.Fatalf("❌ Error checking if bucket exists: %v", errBucketExists)
		}
		if !exists {
			log.Fatalf("❌ Bucket %s does not exist and could not be created: %v", config.MinIO_Bucket, err)
		}
	} else {
		log.Printf("✅ Bucket %s created successfully", config.MinIO_Bucket)
	}

	return &minIO{
		bucket: config.MinIO_Bucket,
		client: client,
	}
}

func (m *minIO) UploadFile(ctx context.Context, path string, file *requests.File) (*models.UploadedFile, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, errors.New("Cannot generate uuid")
	}

	_path := fmt.Sprintf("%s/%s.%s", path, id, file.Extension())
	info, err := m.client.PutObject(ctx, m.bucket, _path, file.Reader, file.Size, minio.PutObjectOptions{
		ContentType: file.ContentType,
	})
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: fmt.Sprintf("Failed to upload file to storage : %s", err.Error())})
	}

	return &models.UploadedFile{
		Name: file.Name,
		Path: _path,
		Size: info.Size,
	}, nil
}

func (m *minIO) UploadMultipleFiles(ctx context.Context, path string, files []requests.File) ([]models.UploadedFile, error) {
	uploadedFiles := make([]models.UploadedFile, len(files))

	for i := range len(files) {
		file, err := m.UploadFile(ctx, fmt.Sprintf("%s/%s", path, files[i].Name), &files[i])
		if err != nil {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: fmt.Sprintf("Failed to upload file %s: %v", files[i].Name, err)})
		}
		uploadedFiles[i] = *file
	}
	return uploadedFiles, nil
}

func (m *minIO) DeleteFile(ctx context.Context, path string) error {
	return m.client.RemoveObject(ctx, m.bucket, path, minio.RemoveObjectOptions{})
}

func (m *minIO) DeleteManyFiles(ctx context.Context, paths []string) error {
	for _, path := range paths {
		err := m.DeleteFile(ctx, path)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: fmt.Sprintf("Failed to delete file %s: %v", path, err)})
		}
	}
	return nil
}
