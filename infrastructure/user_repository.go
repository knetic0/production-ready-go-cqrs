package infrastructure

import (
	"github.com/knetic0/production-ready-go-cqrs/domain"
	"gorm.io/gorm"
)

type UserRepositoryAdapter struct {
	db *gorm.DB
}

func NewUserRepositoryAdaper(db *gorm.DB) *UserRepositoryAdapter {
	return &UserRepositoryAdapter{db: db}
}

func (r *UserRepositoryAdapter) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepositoryAdapter) Get(id string) (*domain.User, error) {
	var u domain.User
	if err := r.db.Where("id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepositoryAdapter) List() ([]domain.User, error) {
	var users []domain.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
