package handlers

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rpegorov/go-parser/internal/api"
)

type IndicatorResponse struct {
	Status     string         `json:"status"`
	Added      IndicatorCount `json:"added"`
	Error      string         `json:"error,omitempty"`
	StatusCode int            `json:"statusCode"`
}

type IndicatorCount struct {
	Total int
}

func (h *Handler) ParseIndicators(c *fiber.Ctx) error {
	cookies := h.CookieStore.GetAll()
	if cookies == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "No cookies found, please login first"})
	}

	equipments := h.indicatorService.GetEquipment()
	if len(equipments) == 0 {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "No equipments found, please parse enterprise first"})
	}

	var totalAdded int
	for _, equipment := range equipments {
		body, err := api.GetIndicatorsByEquipmentIds(equipment, cookies)
		if err != nil {
			log.Printf(`Ошибка обработки индикаторов для оборудования: %s`, err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse indicators for equipment"})
		}
		added, err := h.indicatorService.ParseIndicators(body, equipment)
		if err != nil {
			log.Printf("Ошибка обработки индикаторов: %v", err)
			return c.Status(http.StatusInternalServerError).JSON(IndicatorResponse{
				Status:     "error",
				Error:      err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
		}
		totalAdded += added.Total
	}

	return c.JSON(IndicatorResponse{
		Status:     "success",
		Added:      IndicatorCount{Total: totalAdded},
		StatusCode: http.StatusOK,
	})
}
