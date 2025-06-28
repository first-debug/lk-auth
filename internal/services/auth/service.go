package auth

import (
	"auth-service/internal/domain/models"
	"auth-service/internal/services/jwt"
	"auth-service/internal/services/storage"
	"errors"
)

type AuthServiceImpl struct {
	JWTService jwt.JWTService

	BlackListStorage storage.BlackListStorage
	JWTStorage       storage.JWTStorage
	UserStorage      storage.UserStorage
}

func (s *AuthServiceImpl) Login(email string, passwordHash []byte) (string, string, error) {
	version, role, err := s.UserStorage.Login(email, passwordHash)
	if err != nil {
		return "", "", err
	}
	if version == -1 {
		return "", "", errors.New("incorrect email and password")
	}
	if role == "" {
		return "", "", errors.New("incorrect email and password")
	}

	accessToken, err := s.JWTService.CreateAccessToken(
		models.User{
			Email:   email,
			Version: version,
			Role:    role,
		})
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.JWTService.CreateRefreshToken(
		models.User{
			Email:   email,
			Version: version,
			Role:    role,
		})
	if err != nil {
		return "", "", err
	}

	err = s.JWTStorage.AddPair(accessToken, refreshToken)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthServiceImpl) Refresh(refreshToken string) (string, string, error) {
	// Поиск в чёрном списке
	ok, err := s.BlackListStorage.IsAllowed(refreshToken)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", errors.New("token blocked")
	}

	user, err := s.JWTService.GetUserInfo(refreshToken)
	if err != nil {
		return "", "", err
	}

	ok, err = s.UserStorage.IsVersionValid(user.Email, user.Version)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", errors.New("version is invalid")
	}

	newAccessToken, err := s.JWTService.CreateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.JWTService.CreateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	relatedAccess, err := s.JWTStorage.GetAccessByRefresh(refreshToken)
	if err != nil {
		return "", "", err
	}

	err = s.BlackListStorage.AddTokens(refreshToken, relatedAccess)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
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
	// TODO: добавить проверки новых полей в payload'е токена

	return true, nil
}

func (s *AuthServiceImpl) Logout(tokens ...string) error {
	return s.BlackListStorage.AddTokens(tokens...)
}
