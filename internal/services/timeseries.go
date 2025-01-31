package services

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/rpegorov/go-parser/internal/api"
	"github.com/rpegorov/go-parser/internal/db"
	"github.com/rpegorov/go-parser/internal/utils"
	"gorm.io/gorm"
)

type TimeSeries struct {
	IndicatorID int    `gorm:"not null"`
	EquipmentID int    `gorm:"not null"`
	DateTime    string `gorm:"not null"`
	Value       string
}

type TimeSeriesServiceImpl struct {
	dbpg *gorm.DB
	dbch *gorm.DB
}

func NewTimeseriesService(dbpg *gorm.DB, dbch *gorm.DB) *TimeSeriesServiceImpl {
	return &TimeSeriesServiceImpl{
		dbpg: dbpg,
		dbch: dbch,
	}
}

const (
	maxConcurrentRequests = 5      // Максимальное количество параллельных запросов
	requestDelay          = 50     // Задержка между запросами в миллисекундах
	bufferSize            = 500000 // Размер буфера для данных
	chunkSize             = 100000 // Размер чанка для записи в БД
)

func (ts *TimeSeriesServiceImpl) ParseTimeseries(cookies string) error {
	indicators := ts.GetAllIndicators()
	log.Printf("Получено индикаторов: %d", len(indicators))

	dataStart := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	dataEnd := time.Now()
	const apiDateFormat = "2006-01-02T15:04:05.000Z"

	dataChan := make(chan TimeSeries, bufferSize)
	errorChan := make(chan error)

	var wg sync.WaitGroup

	go ts.processDataChunks(dataChan, errorChan)

	for _, indicator := range indicators {
		wg.Add(1)
		go ts.processIndicator(indicator, dataStart, dataEnd, apiDateFormat, cookies, dataChan, &wg)
	}

	wg.Wait()
	close(dataChan)

	// Обработка ошибок
	var errList []error
	for err := range errorChan {
		errList = append(errList, err)
	}

	if len(errList) > 0 {
		return fmt.Errorf("произошли ошибки при обработке данных: %v", errList)
	}

	return nil
}

func (ts *TimeSeriesServiceImpl) processIndicator(
	indicator db.Indicator,
	startTime, endTime time.Time,
	dateFormat, cookies string,
	dataChan chan<- TimeSeries,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	currentStart := startTime

	for currentStart.Before(endTime) {
		periodEnd := currentStart.Add(time.Hour * 24 * 7)

		err := ts.processTimeRange(indicator, currentStart, periodEnd, dateFormat, cookies, dataChan)
		if err != nil {
			log.Printf("Ошибка обработки индикатора %d: %v", indicator.IndicatorID, err)
			return
		}

		currentStart = periodEnd
	}
}

func (ts *TimeSeriesServiceImpl) processTimeRange(
	indicator db.Indicator,
	startTime, endTime time.Time,
	dateFormat, cookies string,
	dataChan chan<- TimeSeries,
) error {
	startStr := startTime.Format(dateFormat)
	endStr := endTime.Format(dateFormat)

	responseData, err := utils.RerformRequest(func() ([]byte, error) {
		return api.GetIndicatorsData(indicator.IndicatorID, indicator.EquipmentID, startStr, endStr, cookies)
	})
	if err != nil {
		log.Printf("Ошибка получения данных: %v", err)
	}

	if len(responseData) == 0 {
		return nil
	}

	cleanedData, err := cleanResponseData(responseData)
	if err != nil {
		return err
	}

	var rawResponse struct {
		Data []map[string]any `json:"data"`
	}

	if err := json.Unmarshal(cleanedData, &rawResponse); err != nil {
		return fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	for _, entry := range rawResponse.Data {
		dateTime, ok := entry["dateTime"].(string)
		if !ok {
			log.Printf("Пропущена запись без корректного dateTime: %+v", entry)
			continue
		}

		value := normalizeValue(entry["value"])

		dataChan <- TimeSeries{
			IndicatorID: indicator.IndicatorID,
			EquipmentID: indicator.EquipmentID,
			DateTime:    dateTime,
			Value:       value,
		}
	}

	return nil
}

func (ts *TimeSeriesServiceImpl) processDataChunks(dataChan <-chan TimeSeries, errorChan chan<- error) {
	buffer := make([]TimeSeries, 0, chunkSize)

	for data := range dataChan {
		buffer = append(buffer, data)

		if len(buffer) >= chunkSize {
			if err := ts.saveChunkToDB(buffer); err != nil {
				errorChan <- fmt.Errorf("ошибка сохранения данных: %w", err)
			}
			buffer = buffer[:0]
		}
	}

	if len(buffer) > 0 {
		if err := ts.saveChunkToDB(buffer); err != nil {
			errorChan <- fmt.Errorf("ошибка сохранения финальных данных: %w", err)
		}
	}
	close(errorChan)
}

func (ts *TimeSeriesServiceImpl) saveChunkToDB(data []TimeSeries) error {
	if len(data) == 0 {
		return nil
	}

	return ts.dbch.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&data).Error
	})
}

func (ts *TimeSeriesServiceImpl) GetAllIndicators() []db.Indicator {
	var indicators []db.Indicator
	ts.dbpg.Find(&indicators)
	return indicators
}

func normalizeValue(value any) string {
	switch v := value.(type) {
	case float64:
		return fmt.Sprintf("%.6f", v)
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	case bool:
		return fmt.Sprintf("%t", v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func cleanResponseData(responseData []byte) ([]byte, error) {
	re := regexp.MustCompile(`"\$type":("[^"]*"),?`)
	return re.ReplaceAll(responseData, []byte("")), nil
}
