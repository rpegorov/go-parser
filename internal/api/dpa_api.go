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
	const maxRetries = 3
	externalURL := utils.GoDotEnvVariable("DPA_SERVER") + "/Dashboard/getIndicatorData"

	client := &http.Client{
		Timeout: 60 * time.Second,
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

	// for attempt := 1; attempt <= maxRetries; attempt++ {
	req, err := http.NewRequest("POST", externalURL, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookies)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Попытка завершилась ошибкой: %s", err)
		// time.Sleep(time.Second * time.Duration(attempt*2))
		// continue
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("сервер вернул код: %d, тело ответа: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела ответа: %v", err)
	}

	return body, nil
}

// return nil, fmt.Errorf("все попытки завершились неудачей")
// }
