package jwt

import (
	"lk-auth/internal/domain/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
)

type MockJWTService struct {
	mock.Mock
}

func (s *MockJWTService) CreateAccessToken(user model.User) (string, error) {
	args := s.Called(user)
	return args.String(0), args.Error(1)
}

func (s *MockJWTService) CreateRefreshToken(user model.User) (string, error) {
	args := s.Called(user)
	return args.String(0), args.Error(1)
}

func (s *MockJWTService) GetTokenClaims(token string) (jwt.MapClaims, error) {
	args := s.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(jwt.MapClaims), args.Error(1)
}

func (s *MockJWTService) GetUserInfo(token string) (model.User, error) {
	args := s.Called(token)
	if ret, ok := args.Get(0).(model.User); ok {
		return ret, args.Error(1)
	}
	return model.User{}, args.Error(1)
}

func (s *MockJWTService) GetVersion(token string) (float64, error) {
	args := s.Called(token)
	if f, ok := args.Get(0).(float64); ok {
		return f, args.Error(1)
	}
	return 0, args.Error(1)
}

func (s *MockJWTService) GetEmail(token string) (string, error) {
	args := s.Called(token)
	return args.String(0), args.Error(1)
}

func (s *MockJWTService) GetRole(token string) (string, error) {
	args := s.Called(token)
	return args.String(0), args.Error(1)
}

func (s *MockJWTService) GetType(token string) (string, error) {
	args := s.Called(token)
	return args.String(0), args.Error(1)
}

func (s *MockJWTService) IsTokenValid(token string) (bool, error) {
	args := s.Called(token)
	return args.Bool(0), args.Error(1)
}
