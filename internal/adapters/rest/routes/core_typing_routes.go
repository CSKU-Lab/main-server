package routes

import (
	"net/http"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	typingsession "github.com/CSKU-Lab/main-server/internal/typing_session"
	"github.com/gofiber/fiber/v3"
)

type startTypingSessionRequest struct {
	LabID     string  `json:"lab_id"`
	SectionID *string `json:"section_id"`
	CourseID  *string `json:"course_id"`
}

func NewCoreTypingRoute(
	router fiber.Router,
	materialService services.MaterialService,
	enrollmentService services.CourseEnrollmentService,
	labSectionService services.LabSectionService,
	secret string,
) {
	materialRouter := router.Group("/materials")

	materialRouter.Post("/:materialID/typing-session", func(c fiber.Ctx) error {
		materialID := c.Params("materialID")
		user := c.Locals("user").(*models.User)

		var req startTypingSessionRequest
		if err := c.Bind().JSON(&req); err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid request body"})
		}

		if req.SectionID == nil && req.CourseID == nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "section_id or course_id is required"})
		}
		if req.LabID == "" {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "lab_id is required"})
		}

		material, err := materialService.GetByID(c.RequestCtx(), materialID)
		if err != nil {
			return err
		}
		if material.Type != "typing" {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Material is not a typing material"})
		}

		if req.SectionID != nil {
			_, err := labSectionService.GetByLabAndSectionID(c.RequestCtx(), req.LabID, *req.SectionID)
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusForbidden, Message: "Not enrolled in this section"})
			}
		} else {
			enrolled, err := enrollmentService.IsEnrolled(c.RequestCtx(), *req.CourseID, user.ID)
			if err != nil {
				return err
			}
			if !enrolled {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusForbidden, Message: "Not enrolled in this course"})
			}
		}

		startedAt := time.Now()
		claims := &typingsession.TokenClaims{
			StudentID:  user.ID,
			MaterialID: materialID,
			LabID:      req.LabID,
			StartedAt:  startedAt,
		}

		token, err := typingsession.GenerateToken(secret, claims)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Failed to generate session token"})
		}

		return c.JSON(fiber.Map{
			"token":      token,
			"started_at": startedAt,
		})
	})
}
