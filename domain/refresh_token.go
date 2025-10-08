package domain

import "time"

type RefreshToken struct {
	Id        string    `json:"-" gorm:"primaryKey;size:36"`
	Token     string    `json:"token" gorm:"not null;size:512"`
	IsUsed    bool      `json:"isUsed" gorm:"not null;default:false"`
	IsRevoked bool      `json:"isRevoked" gorm:"not null;default:false"`
	ExpiresAt time.Time `json:"expiresAt" gorm:"not null"`
	UserId    string    `json:"userId" gorm:"size:36;not null;index"`
	User      User      `json:"user" gorm:"foreignKey:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
