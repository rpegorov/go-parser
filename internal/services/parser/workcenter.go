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

type WorkcenterService interface {
	ParseWorkcenterInfo(cookies string) error
}

type WorkcenterServiceImpl struct {
	db *gorm.DB
}

func NewWorkcenterService(db *gorm.DB) *WorkcenterServiceImpl {
	return &WorkcenterServiceImpl{
		db: db,
	}
}

type WorkCenterInfo struct {
	GanttResult GanttResult `json:"ganttResult"`
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
	IDRecord          int64          `json:"id"`
	Start             string         `json:"start"`
	End               string         `json:"end"`
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

func (s *WorkcenterServiceImpl) ParseWorkcenterInfo(cookies string) error {
	const buffer = 30000
	equipments := s.GetAllEquipmentIds()
	dataStart := time.Date(2023, 11, 14, 0, 0, 0, 0, time.UTC)
	dataEnd := time.Date(2024, 11, 20, 10, 19, 57, 884, time.UTC)
	const apiDateFormat = "2006-01-02T15:04:05.000Z"
	fmt.Print("Start parse\n")
	dataChan := make(chan WorkCenterInfo, buffer)
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

func (s *WorkcenterServiceImpl) getWorkcenterInfo(
	equipment db.Equipment,
	dateStart time.Time,
	dateEnd time.Time,
	apiDateFormat string,
	cookies string,
	dataChan chan<- WorkCenterInfo,
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

func (s *WorkcenterServiceImpl) proccesTimeRange(
	equipment db.Equipment,
	dateStart time.Time,
	dateEnd time.Time,
	apiDateFormat string,
	cookies string,
	dataChan chan<- WorkCenterInfo,
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
	var root []WorkCenterInfo

	if err := json.Unmarshal(responseData, &root); err != nil {
		log.Printf("ошибка парсинга JSON: %v", err)
		return err
	}

	for _, workcenter := range root {
		dataChan <- workcenter
	}
	return nil
}

func (s *WorkcenterServiceImpl) processDataChunks(dataChan <-chan WorkCenterInfo, errorChan chan<- error) {
	const chunk = 100
	buffer := make([]WorkCenterInfo, 0, chunk)

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

func (s *WorkcenterServiceImpl) saveChunkToDB(data []WorkCenterInfo) error {
	if len(data) == 0 {
		return nil
	}
	var records []db.ExtendedWorkCenter
	for _, info := range data {
		records = append(records, mapWorkCenterInfoToExtendedWorkCenter(info)...)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(&records, 5000).Error
	})
}

func (s *WorkcenterServiceImpl) GetAllEquipmentIds() []db.Equipment {
	var equipments []db.Equipment
	s.db.Find(&equipments)
	fmt.Printf("equipments: %v", equipments)
	return equipments
}

func mapWorkCenterInfoToExtendedWorkCenter(info WorkCenterInfo) []db.ExtendedWorkCenter {
	var result []db.ExtendedWorkCenter

	for _, segment := range info.GanttResult.EquipmentSummaryGanttSegments {
		for _, reason := range segment.DowntimeInfo.Reasons {
			result = append(result, db.ExtendedWorkCenter{
				RecordStartDate:       segment.RecordStartDate,
				RecordEndDate:         segment.RecordEndDate,
				ProcessingProgram:     sql.NullString{String: segment.ProcessingProgram.String, Valid: segment.ProcessingProgram.Valid},
				EquipmentID:           segment.EquipmentID,
				MachineStateType:      segment.MachineStateType,
				DownTimeReasons:       reason.DownTimeReasons,
				ReferenceBookReasonID: reason.ReferenceBookReasonID,
				ReasonName:            reason.ReasonName,
				UserName:              reason.UserName,
				OperatorComment:       reason.OperatorComment,
				IDRecord:              segment.IDRecord,
				Start:                 segment.Start,
				End:                   segment.End,
			})
		}
	}
	return result
}
