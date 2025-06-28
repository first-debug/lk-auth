package jwt

import (
	"auth-service/internal/domain/models"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims map[string]any

type JWTService interface {
	CreateAccessToken(user models.User) (string, error)
	CreateRefreshToken(user models.User) (string, error)

	GetUserInfo(token string) (models.User, error)
	GetTokenClaims(tokenString string) (jwt.MapClaims, error)
	GetVersion(token string) (float64, error)
	GetEmail(token string) (string, error)
	GetRole(token string) (string, error)

	IsTokenValid(token string) (bool, error)
}
