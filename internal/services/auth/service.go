package auth

import (
	"auth-service/internal/domain/models"
	"auth-service/internal/services/jwt"
	"auth-service/internal/services/storage"
	"errors"
	"time"
)

type AuthServiceImpl struct {
	BlackListStorage storage.BlackListStorage
	UserStorage      storage.UserStorage
	JWTService       jwt.JWTService
}

func (s *AuthServiceImpl) Login(email string, passwordHash []byte) (string, string, error) {
	version, err := s.UserStorage.Login(email, passwordHash)
	if err != nil {
		return "", "", err
	}
	if version == -1 {
		return "", "", errors.New("incorrect email and password")
	}

	accessToken, err := s.JWTService.CreateToken(
		models.User{
			Email:        email,
			PasswordHash: passwordHash,
			Version:      version,
		}, time.Duration(time.Minute*15))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.JWTService.CreateToken(
		models.User{
			Email:        email,
			PasswordHash: passwordHash,
			Version:      version,
		}, time.Duration(time.Hour*24))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthServiceImpl) Refresh(token string) (string, string, error) {
	// Поиск в чёрном списке
	ok, err := s.BlackListStorage.IsAllowed(token)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", errors.New("token blocked")
	}

	// Проверка версии данных
	version, err := s.JWTService.GetVersion(token)
	if err != nil {
		return "", "", err
	}
	email, err := s.JWTService.GetEmail(token)
	if err != nil {
		return "", "", err
	}

	ok, err = s.UserStorage.IsVersionValid(email, version)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", errors.New("version is invalid")
	}

	accessToken, err := s.JWTService.CreateToken(
		models.User{
			Email:   email,
			Version: version,
		}, time.Duration(time.Minute*15))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.JWTService.CreateToken(
		models.User{
			Email:   email,
			Version: version,
		}, time.Duration(time.Hour*24))
	if err != nil {
		return "", "", err
	}

	err = s.BlackListStorage.AddTokens(token)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Return true if token is valid
func (s *AuthServiceImpl) ValidateToken(token string) (bool, error) {
	ok, err := s.BlackListStorage.IsAllowed(token)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	ok, err = s.JWTService.IsTokenValid(token)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	return true, nil
}

func (s *AuthServiceImpl) Logout(tokens ...string) error {
	return s.BlackListStorage.AddTokens(tokens...)
}
