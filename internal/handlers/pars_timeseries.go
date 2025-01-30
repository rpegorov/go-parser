package handlers

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ParseTimeseries(c *fiber.Ctx) error {
	cookies := h.CookieStore.GetAll()
	if cookies == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "No cookies found, please login first"})
	}

	// Стартуем парсинг асинхронно
	go func() {
		log.Println("Начат парсинг индикаторов")
		if err := h.timeseriesService.ParseTimeseries(cookies); err != nil {
			log.Printf("Ошибка парсинга индикаторов: %v", err)
		} else {
			log.Println("Парсинг индикаторов завершен")
		}
	}()

	// Возвращаем немедленный ответ клиенту
	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Парсинг индикаторов запущен"})
}
