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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела ответа: %v", err)
		return nil, err
	}
	return body, nil
}
