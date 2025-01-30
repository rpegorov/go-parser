package services

type IndicatorService interface {
	ParseIndicators(body []byte, equipmentId int) (IndicatorCount, error)
	GetEquipment() []int
}

type IndicatorCount struct {
	Total int
}
