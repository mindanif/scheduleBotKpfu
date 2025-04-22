package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"scheduleBot/models"
	"strconv"
)

type KfuAPITeacherRepository struct {
	apiURL string
	client *http.Client
}

func NewKfuAPITeacherRepository(apiURL string) TeacherRepository {
	return &KfuAPITeacherRepository{
		apiURL: apiURL,
		client: &http.Client{},
	}
}

func (r *KfuAPITeacherRepository) FindByName(query string) ([]models.Teacher, error) {
	type Employee struct {
		EmployeeID  int    `json:"employee_id"`
		Lastname    string `json:"lastname"`
		Firstname   string `json:"firstname"`
		Middlename  string `json:"middlename"`
		IsTeacher   bool   `json:"is_teacher"`
		Photo       string `json:"photo"`
		Post        string `json:"post"`
		Sex         string `json:"sex"`
		Subdivision string `json:"subdivision"`
		Workphone   string `json:"workphone"`
		Email       string `json:"email"`
	}

	type APIResponse struct {
		Success   bool       `json:"success"`
		Employees []Employee `json:"employees"`
	}
	url := fmt.Sprintf("%s/%s?q=%s", r.apiURL, "employees", query)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result APIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatal("Ошибка разбора JSON::", err)
		return nil, err
	}
	if !result.Success {
		log.Fatal("Ошибка в апи кфу")
		return nil, fmt.Errorf("")
	}
	teachers := make([]models.Teacher, len(result.Employees))
	for i, t := range result.Employees {
		teachers[i] = models.Teacher{
			ID:       strconv.Itoa(t.EmployeeID),
			FullName: t.Lastname + " " + t.Firstname + " " + t.Middlename,
		}
	}
	return teachers, nil
}

func (r *KfuAPITeacherRepository) GetByID(id string) (models.Teacher, error) {
	//TODO: РЕАЛИЗОВАТЬ ПОЛУЧЕНИЕ ФИО ПРЕПОДА ПО ID (ВЕРОЯТНЕЕ ВСЕГО САМОМУ ХРАНИТЬ В БД)
	return models.Teacher{
		ID:       id,
		FullName: "ЗАГЛУШКА",
	}, nil
}
