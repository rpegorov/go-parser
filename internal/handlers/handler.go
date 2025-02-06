package handlers

import (
	"github.com/rpegorov/go-parser/internal/services"
	"github.com/rpegorov/go-parser/internal/services/ml"
	"github.com/rpegorov/go-parser/internal/services/parser"
	"github.com/rpegorov/go-parser/internal/utils"
)

type Handler struct {
	enterpriseService parser.EnterpriseService
	healthService     services.HealthService
	indicatorService  parser.IndicatorService
	timeseriesService parser.TimeseriesService
	MLService         ml.MLService
	workCenterService parser.WorkcenterService
	CookieStore       *utils.CookieStore
}

func New(
	enterpriseService parser.EnterpriseService,
	healthService services.HealthService,
	indicatorService parser.IndicatorService,
	timeseriesService parser.TimeseriesService,
	MLService ml.MLService,
	workCenterService parser.WorkcenterService,
	cookieStore *utils.CookieStore,
) *Handler {
	return &Handler{
		enterpriseService: enterpriseService,
		healthService:     healthService,
		indicatorService:  indicatorService,
		timeseriesService: timeseriesService,
		MLService:         MLService,
		workCenterService: workCenterService,
		CookieStore:       cookieStore,
	}
}
