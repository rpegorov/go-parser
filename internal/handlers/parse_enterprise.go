package handlers

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rpegorov/go-parser/internal/api"
	"github.com/rpegorov/go-parser/internal/utils"
)

type ParserResponse struct {
	Status     string         `json:"status"`
	Added      EnterpriseTree `json:"added"`
	Error      string         `json:"error,omitempty"`
	StatusCode int            `json:"statusCode"`
}

type EnterpriseTree struct {
	Enterprises int `json:"enterprises"`
	Sites       int `json:"sites"`
	Departments int `json:"departments"`
	Equipment   int `json:"equipment"`
}

var externalParserURL = utils.GoDotEnvVariable("DPA_SERVER") + "/EnterpriseStructManagement/getStaticTree"

func (h *Handler) ParseEnterprise(c *fiber.Ctx) error {
	cookies := h.CookieStore.GetAll()

	if cookies == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "No cookies found, please login first"})
	}

	body, err := api.GetStaticTree(cookies)
	if err != nil {
		log.Printf("Ошибка запроса к внешнему серверу: %v", err)
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "Failed to connect to external parser"})
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

	return c.JSON(ParserResponse{
		Status:     "success",
		Added:      EnterpriseTree(result),
		StatusCode: http.StatusOK,
	})
}
