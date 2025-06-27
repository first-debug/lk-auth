package jwt

import (
	"auth-service/internal/domain/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidTokenClaims = errors.New("invalid token claims")

type JWTServiceImpl struct {
	SecretKey []byte
}

func (s *JWTServiceImpl) CreateToken(user models.User, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"email":   user.Email,
			"exp":     float64(time.Now().Add(ttl).Unix()),
			"version": user.Version,
		})

	tokenString, err := token.SignedString(s.SecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// IsTokenValid проверяет валидность JWT-токена с помощью jwt.Parse().
//
// Он возвращает true, если токен успешно разобран и подпись корректна,
// и false с ошибкой в противном случае.
// Все проверки выполняются библиотекой go-jwt.
//
// Используйте этот метод, если нужна базовая проверка валидности.
func (s *JWTServiceImpl) IsTokenValid(tokenString string) (bool, error) {
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return s.SecretKey, nil
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *JWTServiceImpl) GetVersion(tokenString string) (float64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return s.SecretKey, nil
	})
	if err != nil {
		return 0, err
	}

	tokenClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrInvalidTokenClaims
	}

	version, ok := tokenClaims["version"].(float64)
	if !ok {
		return 0, ErrInvalidTokenClaims
	}

	return version, nil
}

func (s *JWTServiceImpl) GetEmail(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return s.SecretKey, nil
	})
	if err != nil {
		return "", err
	}

	tokenClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidTokenClaims
	}

	email, ok := tokenClaims["email"].(string)
	if !ok {
		return "", ErrInvalidTokenClaims
	}

	return email, nil
}
