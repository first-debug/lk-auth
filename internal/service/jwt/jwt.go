package jwt

import (
	"lk-auth/internal/domain/model"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims map[string]any

type JWTService interface {
	CreateAccessToken(model.User) (string, error)
	CreateRefreshToken(model.User) (string, error)

	GetTokenClaims(token string) (jwt.MapClaims, error)
	GetUserInfo(token string) (model.User, error)
	GetVersion(token string) (float64, error)
	GetEmail(token string) (string, error)
	GetRole(token string) (string, error)
	GetType(token string) (string, error)

	IsTokenValid(string) (bool, error)
}
