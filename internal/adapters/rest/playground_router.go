package rest

import "github.com/gofiber/fiber/v3"

func NewPlaygroundRouter(router fiber.Router, h *PlaygroundHandler) {
	router.Post("/playground/execute", h.Execute)
}
