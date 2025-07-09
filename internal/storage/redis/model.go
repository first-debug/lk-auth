package redis

import (
	"lk-auth/internal/domain/model"
)

type User struct {
	Email        string `redis:"email"`
	PasswordHash string `redis:"passHash"`
	// TODO: добавить валидацию для поля Role
	// Пример: `validate:"required,oneof=admin user guest"`
	Role    string  `redis:"role"`
	Version float64 `redis:"version"`
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
