package jwt

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/first-debug/lk-auth/internal/domain/models"
	sl "github.com/first-debug/lk-auth/internal/libs/logger"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidTokenClaims = errors.New("invalid token claims")
var ErrUnknownClaimType = errors.New("unknown target type")

type JWTServiceImpl struct {
	SecretKey  []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration

	log *slog.Logger
}

func NewJWTServiceImpl(secretKey []byte, accessTTL, refreshTTL time.Duration, log *slog.Logger) (JWTService, error) {
	if len(secretKey) < 32 {
		return nil, errors.New("a key of 256 bits or larger MUST be used with HS256 as specified on RFC 7518")
	}
	if len(secretKey) > 1024 {
		return nil, errors.New("secret key is too large (maximum 1024 bytes)")
	}
	if accessTTL > refreshTTL {
		return nil, errors.New("accessTTL must be less than refreshTTL")
	}

	if log == nil {
		log = slog.New(slog.NewTextHandler(os.Stdin, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	return &JWTServiceImpl{
		SecretKey:  secretKey,
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
		log:        log,
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
		s.log.Error("cannot create Access token", sl.Err(err))
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
		s.log.Error("cannot create Refresh token", sl.Err(err))
		return "", err
	}
	return tokenString, nil
}

func (s *JWTServiceImpl) GetTokenClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return s.SecretKey, nil
	})
	if err != nil {
		s.log.Error("cannot get token claime", sl.Err(err))
		return nil, err
	}

	tokenClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		s.log.Error("cannot get token claime", sl.Err(ErrInvalidTokenClaims))
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

func (s *JWTServiceImpl) GetVersion(tokenString string) (float64, error) {
	var version float64
	err := s.getClaim(tokenString, "version", &version)

	return version, err
}

func (s *JWTServiceImpl) GetEmail(tokenString string) (string, error) {
	var email string
	err := s.getClaim(tokenString, "email", &email)

	return email, err
}

func (s *JWTServiceImpl) GetRole(tokenString string) (string, error) {
	var role string
	err := s.getClaim(tokenString, "role", &role)

	return role, err
}

func (s *JWTServiceImpl) GetType(tokenString string) (string, error) {
	var userType string
	err := s.getClaim(tokenString, "type", &userType)

	return userType, err
}

func (s *JWTServiceImpl) IsTokenValid(tokenString string) (bool, error) {
	tokenClaims, err := s.GetTokenClaims(tokenString)

	if err != nil {
		s.log.Error("JWT validation failed", sl.Err(err))
		return false, err
	}

	var errFieldsBuilder strings.Builder

	email, ok := tokenClaims["email"].(string)
	if res, _ := regexp.MatchString(
		`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
		email,
	); !ok || !res {
		errFieldsBuilder.WriteString("'email'")
	}

	role, ok := tokenClaims["role"].(string)
	if !ok || role == "" {
		if errFieldsBuilder.Len() == 0 {
			errFieldsBuilder.WriteString(", ")
		}
		errFieldsBuilder.WriteString("'role'")
	}

	userType, ok := tokenClaims["type"].(string)
	if !ok || userType == "" {
		if errFieldsBuilder.Len() == 0 {
			errFieldsBuilder.WriteString(", ")
		}
		errFieldsBuilder.WriteString("'type'")
	}

	version, ok := tokenClaims["version"].(float64)
	if !ok || version < 0 {
		if errFieldsBuilder.Len() == 0 {
			errFieldsBuilder.WriteString(", ")
		}
		errFieldsBuilder.WriteString("'version'")
	}

	if errFieldsBuilder.Len() != 0 {
		s.log.Error("invalid token payload", "error with: ", errFieldsBuilder.String())
		return false, errors.New("faild to get " + errFieldsBuilder.String())
	}

	return true, nil
}

func (s *JWTServiceImpl) getClaim(tokenString, name string, target any) error {
	tokenClaims, err := s.GetTokenClaims(tokenString)
	if err != nil {
		s.log.Error("cannot get token claime", sl.Err(err))
		return ErrInvalidTokenClaims
	}

	switch t := target.(type) {
	case *float64:
		val, ok := tokenClaims[name].(float64)
		if !ok {
			s.log.Error("cannot get token claime", sl.Err(ErrInvalidTokenClaims))
			return ErrInvalidTokenClaims
		}
		*t = val
	case *string:
		val, ok := tokenClaims[name].(string)
		if !ok {
			s.log.Error("cannot get token claime", sl.Err(ErrInvalidTokenClaims))
			return ErrInvalidTokenClaims
		}
		*t = val
	default:
		var (
			floatType *float64
			stringype *string
		)
		s.log.Error("unknown target type",
			sl.Err(ErrInvalidTokenClaims),
			"input type", fmt.Sprintf("%T", t),
			"valid types", fmt.Sprintf("%T, %T", floatType, stringype),
		)
		return ErrInvalidTokenClaims
	}

	return nil
}
