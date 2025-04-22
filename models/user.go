package models

type User struct {
	ChatID          int64
	SelectedTeacher Teacher
	Registered      bool
}
