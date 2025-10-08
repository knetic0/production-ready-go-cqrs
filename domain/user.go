package domain

type User struct {
	Id            string         `json:"-" gorm:"primaryKey;size:36"`
	FirstName     string         `json:"firstName" gorm:"size:100;not null"`
	LastName      string         `json:"lastName" gorm:"size:100;not null"`
	Email         string         `json:"email" gorm:"uniqueIndex;size:255;not null"`
	Password      string         `json:"-" gorm:"size:255;not null"`
	RefreshTokens []RefreshToken `json:"refreshTokens" gorm:"foreignKey:UserId"`
}
