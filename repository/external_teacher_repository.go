package repository

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"scheduleBot/models"
)

type ExternalTeacherRepository struct {
	apiURL string
	token  string
	client *http.Client
}

func NewExternalTeacherRepository(apiURL, token string) TeacherRepository {
	return &ExternalTeacherRepository{
		apiURL: apiURL,
		token:  token,
		client: &http.Client{},
	}
}

func (r *ExternalTeacherRepository) fetchTeachers() ([]models.Teacher, error) {
	req, err := http.NewRequest("GET", r.apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Charset", "UTF-8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Campus/4.15.0 (ru.dewish.campus; build:99; iOS iOS 18.3.1)")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.token))
	req.Header.Set("Accept-Language", "ru")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Ожидается, что API возвращает JSON-массив объектов с полями "_id" и "name".
	var apiTeachers []struct {
		ID   string `json:"_id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &apiTeachers); err != nil {
		return nil, err
	}

	teachers := make([]models.Teacher, len(apiTeachers))
	for i, t := range apiTeachers {
		teachers[i] = models.Teacher{
			ID:       t.ID,
			FullName: t.Name,
		}
	}
	return teachers, nil
}

func (r *ExternalTeacherRepository) FindByName(query string) ([]models.Teacher, error) {
	teachers, err := r.fetchTeachers()
	if err != nil {
		return nil, err
	}
	var results []models.Teacher
	q := strings.ToLower(query)
	for _, teacher := range teachers {
		if strings.Contains(strings.ToLower(teacher.FullName), q) {
			results = append(results, teacher)
		}
	}
	return results, nil
}

func (r *ExternalTeacherRepository) GetByID(id string) (models.Teacher, error) {
	teachers, err := r.fetchTeachers()
	if err != nil {
		return models.Teacher{}, err
	}
	for _, teacher := range teachers {
		if teacher.ID == id {
			return teacher, nil
		}
	}
	return models.Teacher{}, fmt.Errorf("преподаватель не найден")
}
