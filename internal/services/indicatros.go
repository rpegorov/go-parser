package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/rpegorov/go-parser/internal/db"
	"gorm.io/gorm"
)

type IndicatorData struct {
	IndicatorID   int    `json:"id"`
	IndicatorName string `json:"name"`
}

type IndicatorServiceImpl struct {
	db *gorm.DB
}

func NewIndicatorService(db *gorm.DB) *IndicatorServiceImpl {
	return &IndicatorServiceImpl{
		db: db,
	}
}

func (s *IndicatorServiceImpl) ParseIndicators(body []byte, equipmentId int) (IndicatorCount, error) {
	var result IndicatorCount
	var root struct {
		Indicators []IndicatorData `json:"data"`
	}
	if err := json.Unmarshal(body, &root); err != nil {
		log.Printf("Ошибка парсинга JSON: %v", err)
		return IndicatorCount{}, err
	}
	if len(root.Indicators) == 0 {
		return IndicatorCount{}, fmt.Errorf("no indicators data found")
	}

	existingIndicators := s.getExistingIndicators(equipmentId)
	for _, indicatorData := range root.Indicators {
		if _, exists := existingIndicators[indicatorData.IndicatorID]; !exists {
			indicator := db.Indicator{
				IndicatorID:   indicatorData.IndicatorID,
				IndicatorName: indicatorData.IndicatorName,
				EquipmentID:   equipmentId,
			}
			if err := s.db.Create(&indicator).Error; err != nil {
				return result, err
			}
		}
		result.Total++
	}
	return result, nil
}

func (s *IndicatorServiceImpl) getExistingIndicators(equipmentID int) map[int]bool {
	indicatorIDs := make(map[int]bool)

	var indicators []db.Indicator
	s.db.Where("equipment_id", equipmentID).Find(&indicators)
	for _, i := range indicators {
		indicatorIDs[i.IndicatorID] = true
	}

	return indicatorIDs
}

func (s *IndicatorServiceImpl) GetEquipment() []int {
	var equipment []db.Equipment
	var equipmentIDs []int
	s.db.Select("equipment_id").Find(&equipment)
	for _, e := range equipment {
		equipmentIDs = append(equipmentIDs, e.EquipmentID)
	}
	return equipmentIDs
}
