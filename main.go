package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/rpegorov/go-parser/internal/db"
	"github.com/rpegorov/go-parser/internal/handlers"
	"github.com/rpegorov/go-parser/internal/middlewares"
	"github.com/rpegorov/go-parser/internal/routes"
	"github.com/rpegorov/go-parser/internal/services"
	"github.com/rpegorov/go-parser/internal/utils"
)

func main() {
	databases := db.Init()
	cookiesStore := utils.NewCookieStore()
	enterpriceService := services.NewEnterpriseService(databases.PostgresDB)
	healthService := services.NewHealthService(databases.PostgresDB, databases.ClickHouseDB)
	indicatorService := services.NewIndicatorService(databases.PostgresDB)

	app := fiber.New(fiber.Config{Prefork: true})
	app.Use(middlewares.CookieMiddleware())
	h := handlers.New(
		enterpriceService,
		healthService,
		indicatorService,
		cookiesStore,
	)

	routes.RegisterRoutes(app, h)

	if err := app.Listen(utils.GoDotEnvVariable("SERVER_PORT")); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
