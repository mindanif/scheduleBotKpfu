package schedule

import (
	"fmt"
	"strings"
	"time"
)

type ScheduleResponse interface {
	FormatForDay(day string) string
	SetSubjects(subjects []Subject)
	GetSubjects() []Subject
}

type ScheduleResponseKFU struct {
	Success  bool      `json:"success"`
	Subjects []Subject `json:"subjects"`
}

type Subject struct {
	ID                    string `json:"id"`
	Semester              int    `json:"semester"`
	Year                  int    `json:"year"`
	SubjectName           string `json:"subject_name"`
	SubjectID             int    `json:"subject_id"`
	StartDaySchedule      string `json:"start_day_schedule"`
	FinishDaySchedule     string `json:"finish_day_schedule"`
	DayWeekSchedule       int    `json:"day_week_schedule"`
	TypeWeekSchedule      int    `json:"type_week_schedule"`
	NoteSchedule          string `json:"note_schedule"`
	TotalTimeSchedule     string `json:"total_time_schedule"`
	BeginTimeSchedule     string `json:"begin_time_schedule"`
	EndTimeSchedule       string `json:"end_time_schedule"`
	TeacherID             int    `json:"teacher_id"`
	TeacherLastname       string `json:"teacher_lastname"`
	TeacherFirstname      string `json:"teacher_firstname"`
	TeacherMiddlename     string `json:"teacher_middlename"`
	NumAuditoriumSchedule string `json:"num_auditorium_schedule"`
	BuildingName          string `json:"building_name"`
	BuildingID            string `json:"building_id"`
	GroupList             string `json:"group_list"`
	SubjectKindName       string `json:"subject_kind_name"`
}

func (sr *ScheduleResponseKFU) FormatForDay(day string) string {
	var sb strings.Builder

	weekday, err := dayStringToInt(day)
	if err != nil {
		return "Ошибка: " + err.Error()
	}
	actualDate, err := computeActualDateForWeekday(weekday)
	if err != nil {
		return "Ошибка вычисления даты: " + err.Error()
	}

	filtered := []Subject{}
	const layout = "02.01.06" // формат дат
	for _, subj := range sr.Subjects {

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

	if len(filtered) == 0 {
		return "Нет занятий на данный день."
	}

	for _, subj := range filtered {
		sb.WriteString(fmt.Sprintf("*Предмет:* %s\n", subj.SubjectName))
		sb.WriteString(fmt.Sprintf("*Время:* %s\n", subj.TotalTimeSchedule))
		sb.WriteString(fmt.Sprintf("*Аудитория:* %s\n", subj.NumAuditoriumSchedule))
		sb.WriteString(fmt.Sprintf("*Здание:* %s\n", subj.BuildingName))
		sb.WriteString(fmt.Sprintf("*Группа:* %s\n", subj.GroupList))
		sb.WriteString(fmt.Sprintf("*Преподаватель:* %s %s %s\n", subj.TeacherLastname, subj.TeacherFirstname, subj.TeacherMiddlename))
		sb.WriteString("────────────────────\n")
	}
	return sb.String()
}

func (sr *ScheduleResponseKFU) GetSubjects() []Subject {
	return sr.Subjects
}

func (sr *ScheduleResponseKFU) SetSubjects(subjects []Subject) {
	sr.Subjects = subjects
}
