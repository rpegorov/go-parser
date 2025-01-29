package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rpegorov/go-parser/internal/utils"
)

type LoginRequest struct {
	UserName   string `json:"UserName" validate:"required"`
	Password   string `json:"Password" validate:"required"`
	RememberMe bool   `json:"RememberMe"`
}

var externalLoginURL = utils.GoDotEnvVariable("DPA_SERVER") + "/Account/Login"

func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to marshal request"})
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	httpReq, err := http.NewRequest("POST", externalLoginURL, bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create request"})
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "External server unavailable"})
	}
	defer resp.Body.Close()

	for _, cookie := range resp.Cookies() {
		h.CookieStore.Set(cookie.Name, cookie.Value)
		// log.Printf("Сохранен cookie: %s=%s", cookie.Name, cookie.Value)
	}

	return c.Status(resp.StatusCode).SendString("Login processed")
}
