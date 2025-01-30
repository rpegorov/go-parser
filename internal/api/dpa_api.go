package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/rpegorov/go-parser/internal/utils"
)

func GetStaticTree(cookie string) (data []byte, err error) {
	var externalUrl = utils.GoDotEnvVariable("DPA_SERVER") + "/EnterpriseStructManagement/getStaticTree"
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	req, err := http.NewRequest("GET", externalUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка запроса к внешнему серверу: %v", err)
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела ответа: %v", err)
		return nil, err
	}
	return body, nil
}

func GetIndicatorsByEquipmentIds(equipmentId int, cookies string) (data []byte, err error) {
	var externalURL = utils.GoDotEnvVariable("DPA_SERVER") + "/Indicator/GetByEquipmentIds"
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	payLoad := fmt.Sprintf(`{"EquipmentIds": [%d]}`, equipmentId)
	req, err := http.NewRequest("POST", externalURL, strings.NewReader(payLoad))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookies)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка запроса к внешнему серверу: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела ответа: %v", err)
		return nil, err
	}
	return body, nil
}

func GetIndicatorsData(
	indicatorId int,
	equipmentId int,
	periodStart string,
	periodEnd string,
	cookies string,
) (data []byte, err error) {
	var externalURL = utils.GoDotEnvVariable("DPA_SERVER") + "/Dashboard/getIndicatorData"

	client := &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10, // Уменьшаем количество простаивающих соединений
			MaxConnsPerHost:     10, // Ограничиваем количество соединений на хост
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   true, // Отключаем keep-alive
			MaxIdleConnsPerHost: 2,
		},
	}
	payload := fmt.Sprintf(`{
        "Indicators": [%d],
        "EquipmentId": %d,
        "ShowUnclassified": true,
        "GroupBy": true,
        "NotTruncateDatePeriod": true,
        "ShowServerState": true,
        "DateTimeFrom": "%s",
        "DateTimeUntil": "%s"
    }`, indicatorId, equipmentId, periodStart, periodEnd)
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("POST", externalURL, strings.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookies)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Connection", "close")

		resp, err := client.Do(req)
		if err != nil {
			if attempt == maxRetries {
				return nil, fmt.Errorf("final attempt failed: %v", err)
			}
			log.Printf("Attempt %d failed: %v. Retrying...", attempt, err)
			time.Sleep(time.Second * time.Duration(attempt*2)) // Увеличивающаяся задержка между попытками
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("server returned status code: %d, body: %s", resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %v", err)
		}

		return body, nil
	}

	return nil, fmt.Errorf("all retry attempts failed")
}
