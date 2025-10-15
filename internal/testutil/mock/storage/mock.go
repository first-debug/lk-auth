package storage

import (
	"context"

	"lk-auth/internal/domain/model"

	"github.com/stretchr/testify/mock"
)

type MockBlackListStorage struct {
	mock.Mock
}

func (s *MockBlackListStorage) AddTokens(tokens ...string) error {
	args := s.Called(tokens)
	return args.Error(0)
}

func (s *MockBlackListStorage) IsAllowed(token string) (bool, error) {
	args := s.Called(token)
	return args.Bool(0), args.Error(1)
}

func (s *MockBlackListStorage) ShutDown(shutDownCtx context.Context) error {
	args := s.Called(shutDownCtx)
	return args.Error(0)
}

type MockUserStorage struct {
	mock.Mock
}

func (s *MockUserStorage) Login(email, passwordHash string) (float64, string, error) {
	args := s.Called(email, passwordHash)
	if f, ok := args.Get(0).(float64); ok {
		return f, args.String(1), args.Error(2)
	}
	return 0, args.String(1), args.Error(2)
}

func (s *MockUserStorage) IsVersionValid(email string, version float64) (bool, error) {
	args := s.Called(email, version)
	return args.Bool(0), args.Error(1)
}

func (s *MockUserStorage) AddUser(user *model.User) error {
	args := s.Called(user)
	return args.Error(0)
}

func (s *MockUserStorage) ShutDown(shutDownCtx context.Context) error {
	args := s.Called(shutDownCtx)
	return args.Error(0)
}

type MockJWTStorage struct {
	mock.Mock
}

func (s *MockJWTStorage) AddPair(access string, refresh string) error {
	args := s.Called(access, refresh)
	return args.Error(0)
}

func (s *MockJWTStorage) GetAccessByRefresh(refresh string) (string, error) {
	args := s.Called(refresh)
	return args.String(0), args.Error(1)
}

func (s *MockJWTStorage) ShutDown(shutDownCtx context.Context) error {
	args := s.Called(shutDownCtx)
	return args.Error(0)
}
