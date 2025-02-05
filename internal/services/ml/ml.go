package ml

import (
	"database/sql"
	"strconv"

	"github.com/rpegorov/go-parser/internal/db"
	"gorm.io/gorm"
)

type MLService interface {
	GetByDataRangeAndEqIdIndId(dataStart, dataEnd, equipment, indicator string) ([]db.TimeSeries, error)
	GetEquipmentTree() ([]EquipmentTree, error)
}

type MLServiceImpl struct {
	dbpg *gorm.DB
	dbch *gorm.DB
}

type EquipmentTree struct {
	EquipmentId   int          `json:"equipment_id"`
	EquipmentName string       `json:"equipment_name"`
	Indicators    []Indicators `json:"indicators"`
}

type Indicators struct {
	IndicatorId   int    `json:"indicator_id"`
	IndicatorName string `json:"indicator_name"`
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

func (s *MLServiceImpl) GetEquipmentTree() ([]EquipmentTree, error) {
	rows, err := s.dbpg.Table("equipment").
		Select("equipment.equipment_id, equipment.equipment_name, indicators.indicator_id, indicators.indicator_name").
		Joins("LEFT JOIN indicators ON indicators.equipment_id = equipment.equipment_id").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	equipmentMap := make(map[int]*EquipmentTree)

	for rows.Next() {
		var eqID int
		var eqName string
		var indID sql.NullInt64
		var indName sql.NullString

		if err := rows.Scan(&eqID, &eqName, &indID, &indName); err != nil {
			return nil, err
		}

		equipment, exists := equipmentMap[eqID]
		if !exists {
			equipment = &EquipmentTree{
				EquipmentId:   eqID,
				EquipmentName: eqName,
				Indicators:    make([]Indicators, 0),
			}
			equipmentMap[eqID] = equipment
		}

		if indID.Valid && indName.Valid {
			equipment.Indicators = append(equipment.Indicators, Indicators{
				IndicatorId:   int(indID.Int64),
				IndicatorName: indName.String,
			})
		}
	}

	// Преобразование map в slice
	result := make([]EquipmentTree, 0, len(equipmentMap))
	for _, equipment := range equipmentMap {
		result = append(result, *equipment)
	}

	return result, nil
}
