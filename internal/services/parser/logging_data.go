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

type LoggingData interface {
	GetLoggingData(cookies string) error
}

type LoggingDataImpl struct {
	db *gorm.DB
}

func NewLoggingData(db *gorm.DB) *LoggingDataImpl {
	return &LoggingDataImpl{
		db: db,
	}
}

type LoggingDataResponse struct {
	Data []Data `json:"data"`
}

type Data struct {
	TimeDuration             string         `json:"timeDuration"`
	TimeDurationDate         string         `json:"timeDurationDate"`
	EquipmentName            string         `json:"equipmentName"`
	EquipmentInventoryNumber string         `json:"equipmentInventoryNumber"`
	Type                     int            `json:"type"`
	TypeText                 string         `json:"typeText"`
	Code                     string         `json:"code"`
	Category                 string         `json:"category"`
	Text                     string         `json:"text"`
	Class                    int            `json:"class"`
	EventLogId               string         `json:"eventLogId"`
	Order                    int            `json:"order"`
	OldEventLogId            sql.NullString `json:"oldEventLogId"`
	EventID                  int            `json:"id"`
	DriverIdentifier         string         `json:"driverIdentifier"`
	TimeStamp                string         `json:"timeStamp"`
	EventStartTime           string         `json:"eventStartTime"`
	EventEndTime             string         `json:"eventEndTime"`
	EventIdentifier          string         `json:"eventIdentifier"`
	Number                   int            `json:"number"`
	StartDate                string         `json:"startDate"`
	EndDate                  string         `json:"endDate"`
}

func (s *LoggingDataImpl) GetLoggingData(cookies string) error {
	const buffer = 30000
	equpments := s.GetAllEquipmentIds()
	dataStart := time.Date(2023, 11, 14, 0, 0, 0, 0, time.UTC)
	dataEnd := time.Date(2024, 11, 20, 10, 19, 57, 884, time.UTC)
	const apiDateFormat = "2006-01-02T15:04:05.000Z"
	dataChan := make(chan LoggingDataResponse, buffer)
	errorChan := make(chan error)

	var wg sync.WaitGroup

	go s.processDataChunks(dataChan, errorChan)

	for _, equipment := range equpments {
		wg.Add(1)
		go s.getLoggingData(equipment, dataStart, dataEnd, apiDateFormat, cookies, dataChan, &wg)
	}

	wg.Wait()
	close(dataChan)

	var errList []error
	for err := range errorChan {
		errList = append(errList, err)
	}

	if len(errList) > 0 {
		return fmt.Errorf("ошибки: %v", errList)
	}

	return nil

}

func (s *LoggingDataImpl) getLoggingData(
	equipment db.Equipment,
	dataStart time.Time,
	dataEnd time.Time,
	apiDateFormat string,
	cookies string,
	dataChan chan<- LoggingDataResponse,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	currentStart := dataStart
	for currentStart.Before(dataEnd) {
		periodEnd := currentStart.Add(time.Hour * 24 * 7)
		err := s.proccesTimeRange(equipment, currentStart, periodEnd, apiDateFormat, cookies, dataChan)
		if err != nil {
			log.Printf("ошибка при обработке данных оборудования : %d: %v", equipment.ID, err)
			continue
		}
		currentStart = periodEnd
	}

}

func (s *LoggingDataImpl) proccesTimeRange(
	equipment db.Equipment,
	dateStart time.Time,
	dateEnd time.Time,
	apiDateFormat string,
	cookies string,
	dataChan chan<- LoggingDataResponse,
) error {

	responseData, err := utils.RerformRequest(func() ([]byte, error) {
		return api.GetLoggingData(equipment.EquipmentID, dateStart.Format(apiDateFormat), dateEnd.Format(apiDateFormat), cookies)
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
	var root LoggingDataResponse

	if err := json.Unmarshal(responseData, &root); err != nil {
		log.Printf("ошибка парсинга JSON: %v", err)
		return err
	}

	dataChan <- root

	return nil
}

func (s *LoggingDataImpl) processDataChunks(dataChan <-chan LoggingDataResponse, errorChan chan<- error) {
	const chunk = 50
	buffer := make([]LoggingDataResponse, 0, chunk)

	for data := range dataChan {
		buffer = append(buffer, data)

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

func (s *LoggingDataImpl) saveChunkToDB(data []LoggingDataResponse) error {
	if len(data) == 0 {
		return nil
	}

	var result []db.LoggingData
	for _, d := range data {
		for _, l := range d.Data {
			loggingRecord := db.LoggingData{
				TimeDuration:             l.TimeDuration,
				TimeDurationDate:         l.TimeDurationDate,
				EquipmentName:            l.EquipmentName,
				EquipmentInventoryNumber: l.EquipmentInventoryNumber,
				Type:                     l.Type,
				TypeText:                 l.TypeText,
				Code:                     l.Code,
				Category:                 l.Category,
				Text:                     l.Text,
				Class:                    l.Class,
				EventLogId:               l.EventLogId,
				Order:                    l.Order,
				OldEventLogId:            l.OldEventLogId,
				EventID:                  l.EventID,
				DriverIdentifier:         l.DriverIdentifier,
				TimeStamp:                l.TimeStamp,
				EventStartTime:           l.EventStartTime,
				EventEndTime:             l.EventEndTime,
				EventIdentifier:          l.EventIdentifier,
				Number:                   l.Number,
				StartDate:                l.StartDate,
				EndDate:                  l.EndDate,
			}
			result = append(result, loggingRecord)
		}
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(result, 1000).Error
	})
}

func (s *LoggingDataImpl) GetAllEquipmentIds() []db.Equipment {
	var equipments []db.Equipment
	s.db.Find(&equipments)
	return equipments
}
