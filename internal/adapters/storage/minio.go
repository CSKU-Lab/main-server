package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/SornchaiTheDev/cs-lab-backend/configs"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/repositories"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minIO struct {
	bucket string
	client *minio.Client
}

func NewMinIOStorage(ctx context.Context, config *configs.Config) repositories.FileRepository {
	client, err := minio.New(config.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.MinIO.AccessKeyID, config.MinIO.SecretAccessKey, ""),
		Secure: config.MinIO.UseSSL,
	})

	if err != nil {
		log.Fatalf("❌ Error creating MinIO client: %v", err)
	}

	err = client.MakeBucket(ctx, config.MinIO.Bucket, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := client.BucketExists(ctx, config.MinIO.Bucket)
		if errBucketExists != nil {
			log.Fatalf("❌ Error checking if bucket exists: %v", errBucketExists)
		}
		if !exists {
			log.Fatalf("❌ Bucket %s does not exist and could not be created: %v", config.MinIO.Bucket, err)
		}
	} else {
		log.Printf("✅ Bucket %s created successfully", config.MinIO.Bucket)
	}

	return &minIO{
		bucket: config.MinIO.Bucket,
		client: client,
	}
}

func (m *minIO) UploadFile(ctx context.Context, path string, file *requests.File) (*models.UploadedFile, error) {
	_path := fmt.Sprintf("%s/%s", path, file.Name)
	info, err := m.client.PutObject(ctx, m.bucket, _path, file.Content, file.Size, minio.PutObjectOptions{
		ContentType: file.MimeType,
	})
	if err != nil {
		return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, fmt.Sprintln("Failed to upload file to storage : ", err.Error()))
	}

	return &models.UploadedFile{
		FileName: file.Name,
		FilePath: fmt.Sprintf("%s/%s", m.bucket, path),
		Size:     info.Size,
	}, nil
}

func (m *minIO) UploadMultipleFiles(ctx context.Context, path string, files []requests.File) ([]models.UploadedFile, error) {
	uploadedFiles := make([]models.UploadedFile, len(files))

	for i := range len(files) {
		file, err := m.UploadFile(ctx, fmt.Sprintf("%s/%s", path, files[i].Name), &files[i])
		if err != nil {
			return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, fmt.Sprintf("Failed to upload file %s: %v", files[i].Name, err))
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
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, fmt.Sprintf("Failed to delete file %s: %v", path, err))
		}
	}
	return nil
}
