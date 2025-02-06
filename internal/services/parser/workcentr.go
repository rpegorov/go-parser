package parser

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/rpegorov/go-parser/internal/api"
	"github.com/rpegorov/go-parser/internal/db"
	"github.com/rpegorov/go-parser/internal/utils"
	"gorm.io/gorm"
)

type WorkcentrService interface {
	ParseWorkcentrInfo(cookies string) error
}

type WorkcentrServiceImpl struct {
	db *gorm.DB
}

func NewWorkcentrService(db *gorm.DB) *WorkcentrServiceImpl {
	return &WorkcentrServiceImpl{
		db: db,
	}
}

type WorkCentrInfo struct {
	GanttResult GanttResult `json:"ganttResult"`
	IDRecord    int64       `json:"id"`
	Start       string      `json:"start"`
	End         string      `json:"end"`
}

type GanttResult struct {
	EquipmentSummaryGanttSegments []EquipmentSummaryGanttSegments `json:"equipmentSummaryGanttSegments"`
}

type EquipmentSummaryGanttSegments struct {
	RecordStartDate   string         `json:"recordStartDate"`
	RecordEndDate     string         `json:"recordEndDate"`
	ProcessingProgram sql.NullString `json:"processingProgram"`
	EquipmentID       int            `json:"equipmentId"`
	MachineStateType  int            `json:"machineStateType"`
	DowntimeInfo      DowntimeInfo   `json:"downtimeInfo"`
}

type DowntimeInfo struct {
	Reasons []Reasons `json:"reasons"`
}

type Reasons struct {
	DownTimeReasons       int    `json:"downtimeReasonId"`
	ReferenceBookReasonID int    `json:"referenceBookReasonId"`
	ReasonName            string `json:"reasonName"`
	UserName              string `json:"userName"`
	OperatorComment       string `json:"operatorComment"`
}

func (s *WorkcentrServiceImpl) ParseWorkcentrInfo(cookies string) error {
	const buffer = 90000
	equipments := s.GetAllEquipmentIds()
	dataStart := time.Date(2023, 11, 14, 0, 0, 0, 0, time.UTC)
	dataEnd := time.Date(2024, 11, 20, 10, 19, 57, 884, time.UTC)
	const apiDateFormat = "2006-01-02T15:04:05.000Z"
	fmt.Print("Start parse\n")
	dataChan := make(chan WorkCentrInfo, buffer)
	errorChan := make(chan error)

	var wg sync.WaitGroup

	go s.processDataChunks(dataChan, errorChan)

	for _, equipment := range equipments {
		wg.Add(1)
		go s.getWorkcenterInfo(equipment, dataStart, dataEnd, apiDateFormat, cookies, dataChan, &wg)
	}

	wg.Wait()
	close(dataChan)

	var errList []error
	for err := range errorChan {
		errList = append(errList, err)
	}

	if len(errList) > 0 {
		return fmt.Errorf("произошли ошибки при обработке данных: %v", errList)
	}

	return nil
}

func (s *WorkcentrServiceImpl) getWorkcenterInfo(
	equipment db.Equipment,
	dateStart time.Time,
	dateEnd time.Time,
	apiDateFormat string,
	cookies string,
	dataChan chan<- WorkCentrInfo,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	currentStart := dateStart

	for currentStart.Before(dateEnd) {
		periodEnd := currentStart.Add(time.Hour * 24 * 7)

		err := s.proccesTimeRange(equipment, currentStart, periodEnd, apiDateFormat, cookies, dataChan)
		if err != nil {
			log.Printf("Ошибка при обработке данных оборудования : %d: %v", equipment.ID, err)
			continue
		}
		currentStart = periodEnd
	}
}

func (s *WorkcentrServiceImpl) proccesTimeRange(
	equipment db.Equipment,
	dateStart time.Time,
	dateEnd time.Time,
	apiDateFormat string,
	cookies string,
	dataChan chan<- WorkCentrInfo,
) error {

	responseData, err := utils.RerformRequest(func() ([]byte, error) {
		return api.GetExtendedWorkCenterSummary(equipment.EquipmentID, dateStart.Format(apiDateFormat), dateEnd.Format(apiDateFormat), cookies)
	})
	if err != nil {
		log.Printf("Ошибка получения данных: %v, оборудования %d, время: %s - %s", err, equipment.EquipmentID, dateStart, dateEnd)
		file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Failed to open log file:", err)
		}
		log.SetOutput(file)
	}
	if len(responseData) == 0 {
		return nil
	}
	var root []WorkCentrInfo

	if err := json.Unmarshal(responseData, &root); err != nil {
		log.Printf("ошибка парсинга JSON: %v", err)
		return err
	}

	for _, workcenter := range root {
		dataChan <- workcenter
	}
	return nil
}

func (s *WorkcentrServiceImpl) processDataChunks(dataChan <-chan WorkCentrInfo, errorChan chan<- error) {
	const chunk = 30000
	buffer := make([]WorkCentrInfo, 0, chunk)

	for data := range dataChan {
		buffer = append(buffer, data)

		fmt.Printf("bufferLen:  %d\n", len(buffer))

		if len(buffer) >= chunk {
			if err := s.saveChunkToDB(buffer); err != nil {
				errorChan <- fmt.Errorf("ошибка сохранения данных: %w", err)
			}
			buffer = buffer[:0]
		}
	}

	if len(buffer) > 0 {
		if err := s.saveChunkToDB(buffer); err != nil {
			errorChan <- fmt.Errorf("ошибка сохранения финальных данных: %w", err)
		}
	}
	close(errorChan)
}

func (s *WorkcentrServiceImpl) saveChunkToDB(data []WorkCentrInfo) error {
	if len(data) == 0 {
		return nil
	}
	if err := s.db.Create(&data).Error; err != nil {
		return err
	}
	return nil
}

func (s *WorkcentrServiceImpl) GetAllEquipmentIds() []db.Equipment {
	var equipments []db.Equipment
	s.db.Find(&equipments)
	fmt.Printf("equipments: %v", equipments)
	return equipments
}
