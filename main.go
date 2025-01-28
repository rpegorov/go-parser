package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/rpegorov/go-parser/internal/db"
	"github.com/rpegorov/go-parser/internal/handlers"
	"github.com/rpegorov/go-parser/internal/utils"
)

func main() {
	databases := db.Init()

	app := fiber.New(fiber.Config{Prefork: false})
	h := handlers.New(databases.PostgresDB, databases.ClickHouseDB)

	app.Get("/api/health", h.HealthCheck)

	if err := app.Listen(utils.GoDotEnvVariable("SERVER_PORT")); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
