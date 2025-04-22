package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"scheduleBot/repository"
	"scheduleBot/schedule"
)

type WebApp struct {
	teacherRepo      repository.TeacherRepository
	scheduleProvider schedule.ScheduleProvider
	templates        *template.Template
}

func NewWebApp(teacherRepo repository.TeacherRepository, scheduleProvider schedule.ScheduleProvider) *WebApp {
	templates := template.Must(template.ParseGlob("web/templates/*.html"))
	return &WebApp{
		teacherRepo:      teacherRepo,
		scheduleProvider: scheduleProvider,
		templates:        templates,
	}
}

func (app *WebApp) Run(addr string) {
	// Главная страница: ожидается, что параметр id передаётся в URL.
	http.HandleFunc("/", app.handleIndex)
	// Новый API-эндпоинт для получения расписания в формате JSON.
	http.HandleFunc("/api/schedule", app.handleAPISchedule)
	log.Printf("Web app запущен по адресу: %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (app *WebApp) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	app.templates.ExecuteTemplate(w, "index.html", nil)
}

func (app *WebApp) handleAPISchedule(w http.ResponseWriter, r *http.Request) {
	teacherID := r.URL.Query().Get("teacher_id")
	day := r.URL.Query().Get("day")
	if teacherID == "" || day == "" {
		http.Error(w, "Отсутствуют необходимые параметры", http.StatusBadRequest)
		return
	}
	scheduleResp, err := app.scheduleProvider.GetSchedule(teacherID)
	if err != nil {
		http.Error(w, "Ошибка получения расписания: "+err.Error(), http.StatusInternalServerError)
		return
	}
	subjects, err := filterSubjectsForDay(scheduleResp, day)
	if err != nil {
		http.Error(w, "Ошибка обработки расписания: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		TeacherID string             `json:"teacher_id"`
		Day       string             `json:"day"`
		Subjects  []schedule.Subject `json:"subjects"`
	}{
		TeacherID: teacherID,
		Day:       day,
		Subjects:  subjects,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func filterSubjectsForDay(sr schedule.ScheduleResponse, day string) ([]schedule.Subject, error) {
	weekday, err := dayStringToInt(day)
	if err != nil {
		return nil, fmt.Errorf("Ошибка: %v", err)
	}
	actualDate, err := computeActualDateForWeekday(weekday)
	if err != nil {
		return nil, fmt.Errorf("Ошибка вычисления даты: %v", err)
	}
	filtered := []schedule.Subject{}
	const layout = "02.01.06" // формат дат, например "11.03.25"
	var subjects []schedule.Subject
	subjects = sr.GetSubjects()
	for _, subj := range subjects {
		if subj.DayWeekSchedule != weekday {
			continue
		}
		if subj.StartDaySchedule != "" && subj.FinishDaySchedule != "" {
			start, err1 := time.Parse(layout, subj.StartDaySchedule)
			finish, err2 := time.Parse(layout, subj.FinishDaySchedule)
			if err1 != nil || err2 != nil {
				continue
			}
			if actualDate.Before(start) || actualDate.After(finish) {
				continue
			}
		}
		filtered = append(filtered, subj)
	}
	return filtered, nil
}

func dayStringToInt(day string) (int, error) {
	mapping := map[string]int{
		"понедельник": 1,
		"вторник":     2,
		"среда":       3,
		"четверг":     4,
		"пятница":     5,
		"суббота":     6,
		"воскресенье": 7,
	}
	if val, ok := mapping[strings.ToLower(day)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("неверный день недели")
}

func computeActualDateForWeekday(weekday int) (time.Time, error) {
	now := time.Now()
	offset := (weekday - int(now.Weekday()) + 7) % 7
	if offset == 0 {
		offset = 7
	}
	return now.AddDate(0, 0, offset), nil
}
