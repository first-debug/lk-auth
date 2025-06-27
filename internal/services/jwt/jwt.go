package jwt

import (
	"auth-service/internal/domain/models"
	"time"
)

type JWTService interface {
	CreateToken(user models.User, ttl time.Duration) (string, error)
	IsTokenValid(token string) (bool, error)
	GetVersion(token string) (float64, error)
	GetEmail(token string) (string, error)
}
