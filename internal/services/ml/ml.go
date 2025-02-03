package ml

import (
	"strconv"

	"github.com/rpegorov/go-parser/internal/db"
	"gorm.io/gorm"
)

type MLService interface {
	GetByDataRangeAndEqIdIndId(dataStart, dataEnd, equipment, indicator string) ([]db.TimeSeries, error)
}

type MLServiceImpl struct {
	dbpg *gorm.DB
	dbch *gorm.DB
}

func NewMLService(dbpg *gorm.DB, dbch *gorm.DB) *MLServiceImpl {
	return &MLServiceImpl{
		dbpg: dbpg,
		dbch: dbch,
	}
}

func (s *MLServiceImpl) GetByDataRangeAndEqIdIndId(dataStart, dataEnd, equipment, indicator string) ([]db.TimeSeries, error) {
	equipmentId, err := strconv.Atoi(equipment)
	if err != nil {
		return nil, err
	}
	indicatorId, err := strconv.Atoi(indicator)
	if err != nil {
		return nil, err
	}

	var results []db.TimeSeries
	err = s.dbch.Where("date_time BETWEEN ? AND ? AND equipment_id = ? AND indicator_id = ?",
		dataStart, dataEnd, equipmentId, indicatorId).Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
