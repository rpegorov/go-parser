package ml

import (
	"database/sql"
	"strconv"

	"github.com/rpegorov/go-parser/internal/db"
	"gorm.io/gorm"
)

type MLService interface {
	GetByDataRangeAndEqIdIndId(dataStart, dataEnd, equipment, indicator string) ([]db.TimeSeries, error)
	GetEquipmentTree() ([]equipmentTree, error)
	GetEquipmentById(equipmentId string) ([]db.Equipment, error)
	GetWorkCentrInfoById(equipmentId string) ([]db.ExtendedWorkCenter, error)
	GetWorkCentrInfoByIdAndDate(equipmentId, dateStart, dateEnd string) ([]db.ExtendedWorkCenter, error)
}

type MLServiceImpl struct {
	dbpg *gorm.DB
	dbch *gorm.DB
}

type equipmentTree struct {
	EquipmentId   int          `json:"equipment_id"`
	EquipmentName string       `json:"equipment_name"`
	Indicators    []indicators `json:"indicators"`
}

type indicators struct {
	IndicatorId   int    `json:"indicator_id"`
	IndicatorName string `json:"indicator_name"`
}

func NewMLService(dbpg *gorm.DB, dbch *gorm.DB) *MLServiceImpl {
	return &MLServiceImpl{
		dbpg: dbpg,
		dbch: dbch,
	}
}

func (s *MLServiceImpl) GetByDataRangeAndEqIdIndId(dateStart, dateEnd, equipment, indicator string) ([]db.TimeSeries, error) {
	equipmentId, err := strconv.Atoi(equipment)
	indicatorId, err := strconv.Atoi(indicator)
	if err != nil {
		return nil, err
	}

	var results []db.TimeSeries
	err = s.dbch.Where("date_time BETWEEN ? AND ? AND equipment_id = ? AND indicator_id = ?",
		dateStart, dateEnd, equipmentId, indicatorId).Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *MLServiceImpl) GetEquipmentTree() ([]equipmentTree, error) {
	rows, err := s.dbpg.Table("equipment").
		Select("equipment.equipment_id, equipment.equipment_name, indicators.indicator_id, indicators.indicator_name").
		Joins("LEFT JOIN indicators ON indicators.equipment_id = equipment.equipment_id").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	equipmentMap := make(map[int]*equipmentTree)

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
			equipment = &equipmentTree{
				EquipmentId:   eqID,
				EquipmentName: eqName,
				Indicators:    make([]indicators, 0),
			}
			equipmentMap[eqID] = equipment
		}

		if indID.Valid && indName.Valid {
			equipment.Indicators = append(equipment.Indicators, indicators{
				IndicatorId:   int(indID.Int64),
				IndicatorName: indName.String,
			})
		}
	}

	result := make([]equipmentTree, 0, len(equipmentMap))
	for _, equipment := range equipmentMap {
		result = append(result, *equipment)
	}

	return result, nil
}

func (s *MLServiceImpl) GetEquipmentById(equipmentId string) ([]db.Equipment, error) {
	var id, err = strconv.Atoi(equipmentId)
	if err != nil {
		return nil, err
	}
	var result = []db.Equipment{}
	err = s.dbpg.Where("equipment.equipment_id = ?", id).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *MLServiceImpl) GetWorkCentrInfoById(equipmentId string) ([]db.ExtendedWorkCenter, error) {
	var id, err = strconv.Atoi(equipmentId)
	if err != nil {
		return nil, err
	}
	var result = []db.ExtendedWorkCenter{}
	err = s.dbpg.Where("equipment_id = ?", id).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *MLServiceImpl) GetWorkCentrInfoByIdAndDate(equipmentId, dateStart, dateEnd string) ([]db.ExtendedWorkCenter, error) {
	var id, err = strconv.Atoi(equipmentId)
	if err != nil {
		return nil, err
	}
	var result = []db.ExtendedWorkCenter{}
	err = s.dbpg.Where("record_start_date BETWEEN ? AND ? AND equipment_id = ?", dateEnd, dateStart, id).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}
