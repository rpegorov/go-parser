package services

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rpegorov/go-parser/internal/api"
	"github.com/rpegorov/go-parser/internal/db"
	"gorm.io/gorm"
)

type TimeseriesData struct {
	Data []Data `json:"data"`
}

type Data struct {
	DateTime string `json:"dateTime"`
	Value    any    `json:"value"`
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

type RequestManager struct {
	semaphore chan struct{}
	delay     time.Duration
}

func NewRequestManager(maxConcurrent int, delay time.Duration) *RequestManager {
	return &RequestManager{
		semaphore: make(chan struct{}, maxConcurrent),
		delay:     delay,
	}
}

func (rm *RequestManager) Execute(f func() error) error {
	rm.semaphore <- struct{}{} // Получаем разрешение
	defer func() {
		<-rm.semaphore       // Освобождаем разрешение
		time.Sleep(rm.delay) // Задержка между запросами
	}()
	return f()
}

const (
	maxConcurrentRequests = 5       // Максимальное количество параллельных запросов
	requestDelay          = 100     // Задержка между запросами в миллисекундах
	bufferSize            = 1000000 // Размер буфера для данных
	chunkSize             = 100000  // Размер чанка для записи в БД
)

func (ts *TimeSeriesServiceImpl) ParseTimeseries(cookies string) error {
	indicators := ts.GetAllIndicators()
	log.Printf("Получено индикаторов: %d", len(indicators))

	dataStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dataEnd := time.Now()
	const apiDateFormat = "2006-01-02T15:04:05.000Z"

	reqManager := NewRequestManager(maxConcurrentRequests, requestDelay*time.Millisecond)

	// Каналы для обработки данных
	dataChan := make(chan TimeseriesData, bufferSize)
	errorChan := make(chan error, len(indicators))

	var wg sync.WaitGroup

	// Запуск обработчика данных
	processWg := sync.WaitGroup{}
	processWg.Add(1)
	go func() {
		defer processWg.Done()
		buffer := make([]TimeseriesData, 0, chunkSize)

		for data := range dataChan {
			buffer = append(buffer, data)

			if len(buffer) >= chunkSize {
				if err := ts.saveChunkToDB(buffer); err != nil {
					errorChan <- fmt.Errorf("ошибка сохранения данных: %w", err)
				}
				buffer = buffer[:0]
			}
		}

		// Сохраняем оставшиеся данные
		if len(buffer) > 0 {
			if err := ts.saveChunkToDB(buffer); err != nil {
				errorChan <- fmt.Errorf("ошибка сохранения финальных данных: %w", err)
			}
		}
	}()

	// Обработка индикаторов
	for _, indicator := range indicators {
		wg.Add(1)
		go func(ind db.Indicator) {
			defer wg.Done()

			currentStart := dataStart
			for currentStart.Before(dataEnd) {
				periodEnd := currentStart.Add(time.Hour * 24 * 7)

				err := reqManager.Execute(func() error {
					return ts.processTimeRange(ind, currentStart, periodEnd, apiDateFormat, cookies, dataChan)
				})

				if err != nil {
					errorChan <- fmt.Errorf("ошибка обработки индикатора %d: %w", ind.IndicatorID, err)
					return
				}

				currentStart = periodEnd
			}
		}(indicator)
	}

	// Ожидание завершения всех горутин
	go func() {
		wg.Wait()
		close(dataChan)
	}()

	// Ожидание завершения обработки данных
	processWg.Wait()
	close(errorChan)

	// Проверка ошибок
	var errList []error
	for err := range errorChan {
		errList = append(errList, err)
	}

	if len(errList) > 0 {
		return fmt.Errorf("произошли ошибки при обработке данных: %v", errList)
	}

	return nil
}

func (ts *TimeSeriesServiceImpl) processTimeRange(
	indicator db.Indicator,
	startTime time.Time,
	endTime time.Time,
	dateFormat string,
	cookies string,
	dataChan chan<- TimeseriesData,
) error {
	startStr := startTime.Format(dateFormat)
	endStr := endTime.Format(dateFormat)

	responseData, err := api.GetIndicatorsData(
		indicator.IndicatorID,
		indicator.EquipmentID,
		startStr,
		endStr,
		cookies,
	)
	if err != nil {
		return fmt.Errorf("ошибка запроса данных: %w", err)
	}

	if len(responseData) == 0 {
		return nil
	}

	var response TimeseriesData
	if err := json.Unmarshal(responseData, &response); err != nil {
		return fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	dataChan <- response
	return nil
}

func (ts *TimeSeriesServiceImpl) saveChunkToDB(data []TimeseriesData) error {
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
