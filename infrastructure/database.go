package infrastructure

import (
	"fmt"

	"github.com/knetic0/production-ready-go-cqrs/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgreAdapter(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("fatal error on postgre connection: %w", err))
	}

	if err := db.AutoMigrate(&domain.User{}, &domain.RefreshToken{}); err != nil {
		panic(fmt.Errorf("fatal error on postgre migration: %w", err))
	}

	return db
}
