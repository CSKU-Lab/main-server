package rest

import (
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/gofiber/fiber/v3"
)

// NewStorageRouter proxies public objects from the storage bucket. The whole
// bucket is anonymous-readable (mc anonymous set public), so this mirrors the
// previous direct-S3 access and is intentionally unauthenticated — register it
// on the public api group, before the protected middleware.
func NewStorageRouter(router fiber.Router, fileRepo repositories.FileRepository) {
	router.Get("/storage/*", func(c fiber.Ctx) error {
		key := c.Params("*")
		if key == "" {
			return fiber.NewError(http.StatusNotFound, "File not found")
		}

		file, err := fileRepo.GetFile(c.Context(), key)
		if err != nil {
			return err
		}
		defer file.Reader.Close()

		if file.ContentType != "" {
			c.Set("Content-Type", file.ContentType)
		}
		if file.ETag != "" {
			c.Set("ETag", file.ETag)
		}
		c.Set("Cache-Control", "public, max-age=3600")

		return c.SendStream(file.Reader, int(file.Size))
	})
}
