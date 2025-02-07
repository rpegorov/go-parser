package handlers

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ErrorLoggingData(c *fiber.Ctx) error {
	cookies := h.CookieStore.GetAll()
	if cookies == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "No cookies found, please login first"})
	}
	go func() {
		log.Println("Start parsing ErrorLoggingData")
		if err := h.loggindDataService.GetLoggingData(cookies); err != nil {
			log.Printf("Ошибка парсинга информации оборудования: %v", err)
		} else {
			log.Println("Парсинг завершен")
		}
	}()
	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Парсинг информации оборудования запущен"})
}
