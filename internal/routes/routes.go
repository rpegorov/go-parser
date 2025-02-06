package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rpegorov/go-parser/internal/handlers"
)

func RegisterRoutes(app *fiber.App, h *handlers.Handler) {
	api := app.Group("/api/v2")
	api.Get("/health", h.HealthCheck)
	api.Post("/login", h.Login)

	ml := api.Group("/ml")
	ml.Get("/time-series", h.GetByDataRangeAndEqIdIndId)
	ml.Get("/equipment-tree", h.GetEquipmentTree)
	ml.Get("/equipment/:id", h.GetEquipmentById)
	ml.Get("/workcenter-info/:id", h.GetWorkCentrInfoById)
	ml.Get("/workcenter-info", h.GetWorkCentrInfoByIdAndDate)

	parser := api.Group("/parser")
	parser.Get("/enterprise", h.ParseEnterprise)
	parser.Get("/indicators", h.ParseIndicators)
	parser.Get("/timeseries", h.ParseTimeseries)
	parser.Get("/workcentr", h.ParseWorkcenter)

}
