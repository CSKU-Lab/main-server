package rest

import (
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest/routes"
	"github.com/CSKU-Lab/queue"
	"github.com/gofiber/fiber/v3"
)

type InternalRouter struct {
	Router          fiber.Router
	InternalToken   string
	CourseService   services.CourseService
	SectionService  services.SectionService
	MaterialService services.MaterialService
	Queue           queue.Queue
}

func NewInternalRouter(r *InternalRouter) {
	routes.NewInternalOGRoutes(r.Router, r.InternalToken, r.CourseService, r.SectionService, r.MaterialService, r.Queue)
}
