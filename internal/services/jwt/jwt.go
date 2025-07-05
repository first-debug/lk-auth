package jwt

import (
	"lk-auth/internal/domain/models"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims map[string]any

type JWTService interface {
	CreateAccessToken(models.User) (string, error)
	CreateRefreshToken(models.User) (string, error)

	GetTokenClaims(token string) (jwt.MapClaims, error)
	GetUserInfo(token string) (models.User, error)
	GetVersion(token string) (float64, error)
	GetEmail(token string) (string, error)
	GetRole(token string) (string, error)
	GetType(token string) (string, error)

	IsTokenValid(string) (bool, error)
}
