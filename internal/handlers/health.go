package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Handler struct {
	PostgresDB   *gorm.DB
	ClickHouseDB *gorm.DB
}

func New(postgresDB *gorm.DB, clickHouseDB *gorm.DB) *Handler {
	return &Handler{
		PostgresDB:   postgresDB,
		ClickHouseDB: clickHouseDB,
	}
}

func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	var postgresStatus string
	if h.PostgresDB != nil {
		pgDB, err := h.PostgresDB.DB()
		if err != nil || pgDB.Ping() != nil {
			postgresStatus = "PostgresDB: Unreachable"
		} else {
			postgresStatus = "PostgresDB: Connected"
		}
	} else {
		postgresStatus = "PostgresDB: Not Configured"
	}

	var clickHouseStatus string
	if h.ClickHouseDB != nil {
		chDB, err := h.ClickHouseDB.DB()
		if err != nil || chDB.Ping() != nil {
			clickHouseStatus = "ClickHouseDB: Unreachable"
		} else {
			clickHouseStatus = "ClickHouseDB: Connected"
		}
	} else {
		clickHouseStatus = "ClickHouseDB: Not Configured"
	}

	status := map[string]string{
		"PostgresDB":   postgresStatus,
		"ClickHouseDB": clickHouseStatus,
	}

	return c.JSON(status)
}
