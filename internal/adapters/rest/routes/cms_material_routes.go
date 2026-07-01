package routes

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/permission"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v3"
)

func NewCMSMaterialRoutes(router fiber.Router, materialService services.MaterialService, materialAssetService services.MaterialAssetService, submissionService services.SubmissionService, inputSubmissionService services.InputSubmissionService, permService permission.Service) {
	materialRouter := router.Group("/courses/:courseID/materials")

	// Create material - instructors and admins only
	materialRouter.Post("/", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), middlewares.Permission(permService).ForCourse("courseID").CanModify(), middlewares.ValidateMiddleware[requests.CreateMaterial](), func(c fiber.Ctx) error {
		req := c.Locals("body").(*requests.CreateMaterial)
		user := c.Locals("user").(*models.User)
		courseID := c.Params("courseID")

		matID, err := materialService.Create(c.RequestCtx(), courseID, user.ID, req)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": matID,
		})
	})

	// List materials - students, instructors, and admins can view
	materialRouter.Get("/", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
		models.STUDENT,
	}), middlewares.Permission(permService).ForCourse("courseID").CanView(), func(c fiber.Ctx) error {
		courseID := c.Params("courseID")
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "10")
		search := c.Query("search", "")
		sortBy := c.Query("sort_by", "name")
		sortOrder := c.Query("sort_order", "desc")

		filterParams := make(map[string]string)
		for key, value := range c.Queries() {
			if strings.Contains(key, "__") {
				filterParams[key] = value
			}
		}

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page"})
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page size"})
		}

		user := c.Locals("user").(*models.User)

		mats, err := materialService.GetPagination(c.RequestCtx(), courseID, user.ID, user.Roles, page, pageSize, search, sortBy, sortOrder, filterParams)
		if err != nil {
			return err
		}

		count, err := materialService.Count(c.RequestCtx(), courseID, user.ID, user.Roles, search, filterParams)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting semesters count"})
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": math.Ceil(float64(count) / float64(pageSize)),
				"total_rows": count,
			},
			"data": mats,
		})
	})

	materialRouter.Post("/fork", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), middlewares.Permission(permService).ForCourse("courseID").CanModify(), middlewares.ValidateMiddleware[requests.ForkMaterial](), func(c fiber.Ctx) error {
		courseID := c.Params("courseID")
		req := c.Locals("body").(*requests.ForkMaterial)
		user := c.Locals("user").(*models.User)

		matID, err := materialService.Fork(c.RequestCtx(), courseID, req.SourceMaterialID, user)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": matID,
		})
	})

	materialRouter.Post("/clone", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), middlewares.Permission(permService).ForCourse("courseID").CanModify(), middlewares.ValidateMiddleware[requests.CloneMaterial](), func(c fiber.Ctx) error {
		courseID := c.Params("courseID")
		req := c.Locals("body").(*requests.CloneMaterial)
		user := c.Locals("user").(*models.User)

		matID, err := materialService.Clone(c.RequestCtx(), courseID, req.SourceMaterialID, user)
		if err != nil {
			return err
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": matID,
		})
	})

	// Get material by ID - students, instructors, and admins can view
	materialRouter.Get("/:id", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
		models.STUDENT,
	}), middlewares.Permission(permService).ForCourse("courseID").CanView(), func(c fiber.Ctx) error {
		courseID := c.Params("courseID")
		id := c.Params("id")
		material, err := materialService.GetByID(c.RequestCtx(), courseID, id)
		if err != nil {
			return err
		}
		return c.JSON(material)
	})

	// Update material - instructors and admins only
	materialRouter.Patch("/:id", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), middlewares.Permission(permService).ForCourse("courseID").CanModify(), middlewares.ValidateMiddleware[requests.BaseUpdateMaterial](), func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		courseID := c.Params("courseID")
		id := c.Params("id")
		req := c.Locals("body").(*requests.BaseUpdateMaterial)
		rawReq := c.Body()

		return materialService.UpdateByID(c.RequestCtx(), courseID, id, req, rawReq, user.ID)
	})

	// Delete material - instructors and admins only
	materialRouter.Delete("/:id", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), middlewares.Permission(permService).ForCourse("courseID").CanModify(), func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		courseID := c.Params("courseID")
		id := c.Params("id")
		return materialService.DeleteByID(c.RequestCtx(), courseID, id, user.ID)
	})

	// List input submissions for a document material - instructors and admins only
	materialRouter.Get("/:id/input-submissions", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), middlewares.Permission(permService).ForCourse("courseID").CanView(), func(c fiber.Ctx) error {
		id := c.Params("id")
		results, err := inputSubmissionService.ListByMaterial(c.RequestCtx(), id)
		if err != nil {
			return err
		}
		return c.JSON(results)
	})

	// Upload asset - instructors and admins only
	materialRouter.Post("/:id/assets", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}), middlewares.Permission(permService).ForCourse("courseID").CanModify(), func(c fiber.Ctx) error {
		filePayload, err := c.FormFile("file")
		if err != nil {
			return err
		}

		file, err := filePayload.Open()
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Failed to open image file"})
		}

		defer file.Close()

		imageFile := &requests.File{
			Name:        filePayload.Filename,
			Size:        filePayload.Size,
			Reader:      file,
			ContentType: filePayload.Header.Get("Content-Type"),
		}

		courseID := c.Params("courseID")
		id := c.Params("id")
		if _, err := materialService.GetByID(c.RequestCtx(), courseID, id); err != nil {
			return err
		}

		fileURL, err := materialAssetService.UploadFile(c.RequestCtx(), id, imageFile)
		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{
			"url": fileURL,
		})
	})
}

// fiber:context-methods migrated
