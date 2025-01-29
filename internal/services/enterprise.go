package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/rpegorov/go-parser/internal/db"
	"gorm.io/gorm"
)

type EnterpriseData struct {
	EnterpriseID   int        `json:"enterpriseId"`
	EnterpriseName string     `json:"text"`
	Items          []SiteData `json:"items"`
}
type SiteData struct {
	SiteID   int              `json:"siteId"`
	SiteName string           `json:"text"`
	Items    []DepartmentData `json:"items"`
}

type DepartmentData struct {
	DepartmentID   int              `json:"departmentId"`
	DepartmentName string           `json:"text"`
	Items          []EquipmentGroup `json:"items"`
}

type EquipmentGroup struct {
	Items []EquipmentData `json:"items"`
}

type EquipmentData struct {
	EquipmentID   int    `json:"equipmentId"`
	EquipmentName string `json:"text"`
}

type EnterpriseServiceImpl struct {
	db *gorm.DB
}

func NewEnterpriseService(db *gorm.DB) *EnterpriseServiceImpl {
	return &EnterpriseServiceImpl{
		db: db,
	}
}

func (s *EnterpriseServiceImpl) ParseAndSaveEnterpriseTree(body []byte) (Result, error) {
	var root struct {
		Items []EnterpriseData `json:"items"`
	}
	if err := json.Unmarshal(body, &root); err != nil {
		log.Printf("Ошибка парсинга JSON: %v", err)
		return Result{}, err
	}

	if len(root.Items) == 0 {
		return Result{}, fmt.Errorf("no enterprise data found")
	}

	var result Result

	existingStructure := s.getExistingStructure()

	for _, enterpriseData := range root.Items {
		if _, exists := existingStructure["enterprise"][enterpriseData.EnterpriseID]; !exists {
			enterprise := db.Enterprise{
				EnterpriseID:   enterpriseData.EnterpriseID,
				EnterpriseName: enterpriseData.EnterpriseName,
			}
			if err := s.db.Create(&enterprise).Error; err != nil {
				log.Printf("Ошибка создания предприятия: %v", err)
				continue
			}
			result.Enterprises++
		}

		for _, siteData := range enterpriseData.Items {
			if _, exists := existingStructure["site"][siteData.SiteID]; !exists {
				site := db.Site{
					SiteID:       siteData.SiteID,
					SiteName:     siteData.SiteName,
					EnterpriseID: enterpriseData.EnterpriseID,
				}
				if err := s.db.Create(&site).Error; err != nil {
					log.Printf("Ошибка создания площадки: %v", err)
					continue
				}
				result.Sites++
			}

			for _, departmentData := range siteData.Items {
				if _, exists := existingStructure["department"][departmentData.DepartmentID]; !exists {
					department := db.Department{
						DepartmentID:   departmentData.DepartmentID,
						DepartmentName: departmentData.DepartmentName,
						SiteID:         siteData.SiteID,
					}
					if err := s.db.Create(&department).Error; err != nil {
						log.Printf("Ошибка создания подразделения: %v", err)
						continue
					}
					result.Departments++
				}

				for _, equipmentGroup := range departmentData.Items {
					for _, equipmentData := range equipmentGroup.Items {
						if _, exists := existingStructure["equipment"][equipmentData.EquipmentID]; !exists {
							equipment := db.Equipment{
								EquipmentID:   equipmentData.EquipmentID,
								EquipmentName: equipmentData.EquipmentName,
								DepartmentID:  departmentData.DepartmentID,
							}
							if err := s.db.Create(&equipment).Error; err != nil {
								log.Printf("Ошибка создания оборудования: %v", err)
								continue
							}
							result.Equipment++
						}
					}
				}
			}
		}
	}

	return result, nil
}

func (s *EnterpriseServiceImpl) getExistingStructure() map[string]map[int]bool {
	structure := map[string]map[int]bool{
		"enterprise": make(map[int]bool),
		"site":       make(map[int]bool),
		"department": make(map[int]bool),
		"equipment":  make(map[int]bool),
	}

	var enterprises []db.Enterprise
	s.db.Select("enterprise_id").Find(&enterprises)
	for _, e := range enterprises {
		structure["enterprise"][e.EnterpriseID] = true
	}

	var sites []db.Site
	s.db.Select("site_id").Find(&sites)
	for _, s := range sites {
		structure["site"][s.SiteID] = true
	}

	var departments []db.Department
	s.db.Select("department_id").Find(&departments)
	for _, d := range departments {
		structure["department"][d.DepartmentID] = true
	}

	var equipment []db.Equipment
	s.db.Select("equipment_id").Find(&equipment)
	for _, e := range equipment {
		structure["equipment"][e.EquipmentID] = true
	}

	return structure
}
