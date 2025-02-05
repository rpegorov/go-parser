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

func (h *Handler) GetEquipmentTree(c *fiber.Ctx) error {
	result, err := h.MLService.GetEquipmentTree()
	if err != nil {
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": result,
	})
}

func (h *Handler) GetEquipmentById(c *fiber.Ctx) error {
	equipmentId := c.Params("id")
	result, err := h.MLService.GetEquipmentById(equipmentId)
	if err != nil {
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": result,
	})
}
