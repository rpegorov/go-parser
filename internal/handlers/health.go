package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	status, err := h.healthService.CheckHealth()
	if err != nil {
		return err
	}

	return c.JSON(status)
}
