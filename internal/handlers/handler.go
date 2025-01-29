package handlers

import (
	"github.com/rpegorov/go-parser/internal/services"
	"github.com/rpegorov/go-parser/internal/utils"
)

type Handler struct {
	enterpriseService services.EnterpriseService
	healthService     services.HealthService
	CookieStore       *utils.CookieStore
}

func New(enterpriseService services.EnterpriseService, healthService services.HealthService, cookieStore *utils.CookieStore) *Handler {
	return &Handler{
		enterpriseService: enterpriseService,
		healthService:     healthService,
		CookieStore:       cookieStore,
	}
}
