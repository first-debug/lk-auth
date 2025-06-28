package jwt

import (
	"auth-service/internal/domain/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidTokenClaims = errors.New("invalid token claims")

type JWTServiceImpl struct {
	SecretKey  []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewJWTServiceImpl(secretKey []byte, accessTTL, refreshTTL time.Duration) (*JWTServiceImpl, error) {
	if len(secretKey) < 32 {
		return nil, errors.New("a key of 256 bits or larger MUST be used with HS256 as specified on RFC 7518")
	}
	if len(secretKey) > 1024 {
		return nil, errors.New("secret key is too large (maximum 1024 bytes)")
	}
	if accessTTL > refreshTTL {
		return nil, errors.New("accessTTL must be less than refreshTTL")
	}

	return &JWTServiceImpl{
		SecretKey:  secretKey,
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	}, nil
}

func (s *JWTServiceImpl) CreateAccessToken(user models.User) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"email":   user.Email,
			"exp":     float64(time.Now().Add(s.AccessTTL).Unix()),
			"role":    user.Role,
			"type":    "access",
			"version": user.Version,
		})

	tokenString, err := token.SignedString(s.SecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *JWTServiceImpl) CreateRefreshToken(user models.User) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"email":   user.Email,
			"exp":     float64(time.Now().Add(s.RefreshTTL).Unix()),
			"role":    user.Role,
			"type":    "refresh",
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
	tokenClaims, err := s.GetTokenClaims(tokenString)
	if err != nil {
		return 0, ErrInvalidTokenClaims
	}

	version, ok := tokenClaims["version"].(float64)
	if !ok {
		return 0, ErrInvalidTokenClaims
	}

	return version, nil
}

func (s *JWTServiceImpl) GetEmail(tokenString string) (string, error) {
	tokenClaims, err := s.GetTokenClaims(tokenString)
	if err != nil {
		return "", ErrInvalidTokenClaims
	}

	email, ok := tokenClaims["email"].(string)
	if !ok {
		return "", ErrInvalidTokenClaims
	}

	return email, nil
}

func (s *JWTServiceImpl) GetRole(tokenString string) (string, error) {
	tokenClaims, err := s.GetTokenClaims(tokenString)
	if err != nil {
		return "", ErrInvalidTokenClaims
	}

	role, ok := tokenClaims["role"].(string)
	if !ok {
		return "", ErrInvalidTokenClaims
	}

	return role, nil
}

func (s *JWTServiceImpl) GetType(tokenString string) (string, error) {
	tokenClaims, err := s.GetTokenClaims(tokenString)
	if err != nil {
		return "", ErrInvalidTokenClaims
	}

	email, ok := tokenClaims["email"].(string)
	if !ok {
		return "", ErrInvalidTokenClaims
	}

	return email, nil
}

func (s *JWTServiceImpl) GetTokenClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return s.SecretKey, nil
	})
	if err != nil {
		return nil, err
	}

	tokenClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidTokenClaims
	}

	return tokenClaims, nil
}

func (s *JWTServiceImpl) GetUserInfo(tokenString string) (models.User, error) {
	user := models.User{}

	if version, err := s.GetVersion(tokenString); err != nil {
		return models.User{}, err
	} else {
		user.Version = version
	}

	if email, err := s.GetEmail(tokenString); err != nil {
		return models.User{}, err
	} else {
		user.Email = email
	}

	if role, err := s.GetRole(tokenString); err != nil {
		return models.User{}, err
	} else {
		user.Role = role
	}

	return user, nil
}
