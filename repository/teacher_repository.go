package repository

import "scheduleBot/models"

type TeacherRepository interface {
	FindByName(query string) ([]models.Teacher, error)
	GetByID(id string) (models.Teacher, error)
}
