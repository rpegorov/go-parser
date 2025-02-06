package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/rpegorov/go-parser/internal/db"
	"github.com/rpegorov/go-parser/internal/handlers"
	"github.com/rpegorov/go-parser/internal/middlewares"
	"github.com/rpegorov/go-parser/internal/routes"
	"github.com/rpegorov/go-parser/internal/services"
	"github.com/rpegorov/go-parser/internal/services/ml"
	"github.com/rpegorov/go-parser/internal/services/parser"
	"github.com/rpegorov/go-parser/internal/utils"
)

func main() {
	databases := db.Init()
	cookiesStore := utils.NewCookieStore()
	enterpriceService := parser.NewEnterpriseService(databases.PostgresDB)
	healthService := services.NewHealthService(databases.PostgresDB, databases.ClickHouseDB)
	indicatorService := parser.NewIndicatorService(databases.PostgresDB)
	timeseriesService := parser.NewTimeseriesService(databases.PostgresDB, databases.ClickHouseDB)
	mlService := ml.NewMLService(databases.PostgresDB, databases.ClickHouseDB)
	workCentrService := parser.NewWorkcentrService(databases.PostgresDB)

	app := fiber.New(fiber.Config{Prefork: false})
	app.Use(middlewares.CookieMiddleware())
	h := handlers.New(
		enterpriceService,
		healthService,
		indicatorService,
		timeseriesService,
		mlService,
		workCentrService,
		cookiesStore,
	)

	routes.RegisterRoutes(app, h)

	if err := app.Listen(utils.GoDotEnvVariable("SERVER_PORT")); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
