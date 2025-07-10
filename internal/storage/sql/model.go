package sql

import (
	"time"

	"lk-auth/internal/domain/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	UUID      uuid.UUID      `gorm:"type:CHAR(36);primaryKey"`
	CreatedAt time.Time      // gorm сам управляет этим полем при операциях INSERT
	UpdatedAt time.Time      // gorm сам управляет этим полем при операциях UPDATE
	DeletedAt gorm.DeletedAt `gorm:"index"` // gorm всегда добаляеть условие `WHERE deleted_at NOT IS NULL` при SELECT-запросах
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.UUID == uuid.Nil {
		m.UUID = uuid.New()
	}
	return
}

type User struct {
	BaseModel

	Email        string  `gorm:"type:VARCHAR(256);not null;uniqueIndex"`
	PasswordHash string  `gorm:"not null"`
	Role         string  `gorm:"not null"`
	Version      float64 `gorm:"not null"`
}

func fromDomain(u *model.User) *User {
	return &User{
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		Version:      u.Version,
	}
}

func (u *User) toDomain() *model.User {
	return &model.User{
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		Version:      u.Version,
	}
}
