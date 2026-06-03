package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/domain/services"
	typingsession "github.com/CSKU-Lab/main-server/internal/typing_session"
	"github.com/gofiber/fiber/v3"
)

type startTypingSessionRequest struct {
	LabID     string  `json:"lab_id"`
	SectionID *string `json:"section_id"`
	CourseID  *string `json:"course_id"`
}

type markTypingStartedRequest struct {
	Token string `json:"token"`
}

func NewCoreTypingRoute(
	router fiber.Router,
	materialService services.MaterialService,
	enrollmentService services.CourseEnrollmentService,
	labSectionService services.LabSectionService,
	typingSubRepo repositories.TypingSubmissionRepository,
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

		material, err := materialService.GetByIDUnscoped(c.RequestCtx(), materialID)
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

	materialRouter.Post("/:materialID/typing-session/mark-started", func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		var req markTypingStartedRequest
		if err := c.Bind().JSON(&req); err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid request body"})
		}
		if req.Token == "" {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "token is required"})
		}

		claims, err := typingsession.VerifyToken(secret, req.Token)
		if err != nil {
			return err
		}
		if claims.StudentID != user.ID {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Session token does not belong to this user"})
		}
		if !claims.TypingStartedAt.IsZero() {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Typing already started"})
		}

		claims.TypingStartedAt = time.Now()
		newToken, err := typingsession.GenerateToken(secret, claims)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Failed to generate session token"})
		}

		return c.JSON(fiber.Map{"token": newToken})
	})

	materialRouter.Get("/:materialID/typing-session/best", func(c fiber.Ctx) error {
		materialID := c.Params("materialID")
		user := c.Locals("user").(*models.User)
		labID := c.Query("lab_id", "")
		sectionID := c.Query("section_id", "")

		if labID == "" || sectionID == "" {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "lab_id and section_id are required"})
		}

		best, err := typingSubRepo.GetBestByUserID(c.RequestCtx(), user.ID, materialID, labID, sectionID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.JSON(fiber.Map{"id": nil})
			}
			return err
		}

		return c.JSON(fiber.Map{
			"id":           best.SubmissionID,
			"raw_wpm":      best.RawWPM,
			"adjusted_wpm": best.AdjustedWPM,
			"error_rate":   best.ErrorRate,
			"duration":     best.Duration,
		})
	})
}
