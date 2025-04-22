package schedule

import (
	_ "bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "strings"
	"time"
)

type KFUProvider struct {
	baseURL string
	client  *http.Client
}

func NewKFUProvider(baseURL string) ScheduleProvider {
	return &KFUProvider{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *KFUProvider) GetSchedule(teacherID string) (ScheduleResponse, error) {
	var sr *ScheduleResponseKFU

	// Формируем URL вида: {baseURL}/employees/{teacherID}/schedule
	url := fmt.Sprintf("%s/employees/%s/schedule", p.baseURL, teacherID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return sr, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		log.Println(err)
		return sr, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return sr, fmt.Errorf("неверный статус ответа: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return sr, err
	}
	// Разбираем JSON-ответ.
	if err := json.Unmarshal(body, &sr); err != nil {
		return sr, err
	}

	return sr, nil
}
