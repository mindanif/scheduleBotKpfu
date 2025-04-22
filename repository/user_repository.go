package repository

import (
	"errors"
	"sync"

	"scheduleBot/models"
)

type UserRepository interface {
	Save(user models.User) error
	Get(chatID int64) (models.User, error)
	Delete(chatID int64) (bool, error)
}

type InMemoryUserRepository struct {
	data  map[int64]models.User
	mutex sync.RWMutex
}

func NewInMemoryUserRepository() UserRepository {
	return &InMemoryUserRepository{
		data: make(map[int64]models.User),
	}
}

func (r *InMemoryUserRepository) Save(user models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.data[user.ChatID] = user
	return nil
}

func (r *InMemoryUserRepository) Get(chatID int64) (models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	user, exists := r.data[chatID]
	if !exists {
		return models.User{}, errors.New("пользователь не найден")
	}
	return user, nil
}
func (r *InMemoryUserRepository) Delete(chatID int64) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	user, exists := r.data[chatID]
	if !exists {
		return false, errors.New("пользователь не найден")
	}
	user.Registered = false
	r.data[chatID] = user
	return true, nil
}
