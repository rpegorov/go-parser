package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetByDataRangeAndEqIdIndId(c *fiber.Ctx) error {
	dataStart := c.Query("dataStart")
	dateEnd := c.Query("dateEnd")
	equipment := c.Query("equipment")
	indicator := c.Query("indicator")

	result, err := h.MLService.GetByDataRangeAndEqIdIndId(dataStart, dateEnd, equipment, indicator)
	if err != nil {
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"data": result,
	})
}
