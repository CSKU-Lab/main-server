package routes

import (
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/gofiber/fiber/v3"
)

type submitInputRequest struct {
	NodeID             string  `json:"node_id"`
	DocumentMaterialID string  `json:"document_material_id"`
	LabID              string  `json:"lab_id"`
	SectionID          *string `json:"section_id"`
	Value              string  `json:"value"`
}

func NewCoreInputRoutes(router fiber.Router, service services.InputSubmissionService, labSectionService services.LabSectionService) {
	inputRouter := router.Group("/input-submissions")

	inputRouter.Post("/", func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		var req submitInputRequest
		if err := c.Bind().JSON(&req); err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid request body"})
		}
		if req.NodeID == "" || req.DocumentMaterialID == "" || req.LabID == "" {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "node_id, document_material_id and lab_id are required"})
		}
		if req.SectionID != nil {
			if err := requireLabAccess(c.RequestCtx(), labSectionService, req.LabID, *req.SectionID, true); err != nil {
				return err
			}
		}

		result, err := service.SubmitInputAnswer(c.RequestCtx(), &services.SubmitInputAnswerInput{
			UserID:             user.ID,
			NodeID:             req.NodeID,
			DocumentMaterialID: req.DocumentMaterialID,
			LabID:              req.LabID,
			SectionID:          req.SectionID,
			Value:              req.Value,
		})
		if err != nil {
			return err
		}

		return c.JSON(result)
	})

	inputRouter.Get("/my-result", func(c fiber.Ctx) error {
		user := c.Locals("user").(*models.User)

		nodeID := c.Query("node_id", "")
		documentMaterialID := c.Query("document_material_id", "")
		labID := c.Query("lab_id", "")

		if nodeID == "" || documentMaterialID == "" || labID == "" {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "node_id, document_material_id and lab_id are required"})
		}

		result, err := service.GetMyLatestResult(c.RequestCtx(), user.ID, nodeID, documentMaterialID, labID)
		if err != nil {
			return err
		}

		return c.JSON(result)
	})
}
