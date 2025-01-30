package services

import "github.com/rpegorov/go-parser/internal/db"

type TimeseriesService interface {
	ParseTimeseries(cookies string) error
	GetAllIndicators() []db.Indicator
}
