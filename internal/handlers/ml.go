package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetByDataRangeAndEqIdIndId(c *fiber.Ctx) error {
	dateStart := c.Query("dateStart")
	dateEnd := c.Query("dateEnd")
	equipment := c.Query("equipment")
	indicator := c.Query("indicator")

	result, err := h.MLService.GetByDataRangeAndEqIdIndId(dateStart, dateEnd, equipment, indicator)
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

func (h *Handler) GetWorkCentrInfoById(c *fiber.Ctx) error {
	equipmentId := c.Params("id")
	result, err := h.MLService.GetWorkCentrInfoById(equipmentId)
	if err != nil {
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": result,
	})
}

func (h *Handler) GetWorkCentrInfoByIdAndDate(c *fiber.Ctx) error {
	equipmentId := c.Query("equipmentId")
	dateStart := c.Query("start")
	dateEnd := c.Query("end")
	result, err := h.MLService.GetWorkCentrInfoByIdAndDate(equipmentId, dateStart, dateEnd)
	if err != nil {
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": result,
	})
}

func (h *Handler) GetLoggingDataByDataRangeAndEqId(c *fiber.Ctx) error {
	equipmentId := c.Params("id")
	dateStart := c.Query("start")
	dateEnd := c.Query("end")
	result, err := h.MLService.GetLoggingDataByDataRangeAndEqId(equipmentId, dateStart, dateEnd)
	if err != nil {
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": result,
	})
}
