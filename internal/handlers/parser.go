package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rpegorov/go-parser/internal/utils"
)

type ParserResponse struct {
	Status     string `json:"status"`
	Added      Result `json:"added"`
	Error      string `json:"error,omitempty"`
	StatusCode int    `json:"statusCode"`
}

type Result struct {
	Enterprises int `json:"enterprises"`
	Sites       int `json:"sites"`
	Departments int `json:"departments"`
	Equipment   int `json:"equipment"`
}

var externalParserURL = utils.GoDotEnvVariable("DPA_SERVER") + "/EnterpriseStructManagement/getStaticTree"

func (h *Handler) ParseEnterprise(c *fiber.Ctx) error {
	cookies := h.CookieStore.GetAll()

	if len(cookies) == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "No cookies found, please login first"})
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	req, err := http.NewRequest("GET", externalParserURL, nil)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create request"})
	}

	cookieHeader := ""
	for name, value := range cookies {
		cookieHeader += fmt.Sprintf("%s=%s; ", name, value)
	}
	cookieHeader = strings.TrimRight(cookieHeader, "; ")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookieHeader)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка запроса к внешнему серверу: %v", err)
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "Failed to connect to external parser"})
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела ответа: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read response body"})
	}

	result, err := h.enterpriseService.ParseAndSaveEnterpriseTree(body)
	if err != nil {
		log.Printf("Ошибка обработки структуры предприятия: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(ParserResponse{
			Status:     "error",
			Error:      err.Error(),
			StatusCode: http.StatusInternalServerError,
		})
	}

	log.Printf("Ответ от внешнего сервера: %v", resp.Status)
	return c.JSON(ParserResponse{
		Status:     "success",
		Added:      Result(result),
		StatusCode: http.StatusOK,
	})
}
