package handlers

import (
	"github.com/rpegorov/go-parser/internal/services"
	"github.com/rpegorov/go-parser/internal/utils"
)

type Handler struct {
	enterpriseService services.EnterpriseService
	healthService     services.HealthService
	indicatorService  services.IndicatorService
	CookieStore       *utils.CookieStore
}

func New(
	enterpriseService services.EnterpriseService,
	healthService services.HealthService,
	indicatorService services.IndicatorService,
	cookieStore *utils.CookieStore,
) *Handler {
	return &Handler{
		enterpriseService: enterpriseService,
		healthService:     healthService,
		indicatorService:  indicatorService,
		CookieStore:       cookieStore,
	}
}
