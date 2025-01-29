package middlewares

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func CookieMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Логируем все cookie для отладки
		cookies := c.Cookies(".")
		if cookies != "" {
			log.Printf("Cookies for request: %s", cookies)
		}
		return c.Next()
	}
}
