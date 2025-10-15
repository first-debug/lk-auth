//go:build integration

package auth_test

import (
	"errors"
	"log/slog"
	"os"
	"testing"

	"lk-auth/internal/domain/model"
	authpkg "lk-auth/internal/service/auth"
	"lk-auth/internal/testutil/mock/jwt"
	"lk-auth/internal/testutil/mock/storage"

	"github.com/stretchr/testify/assert"
)

var (
	correctUser = model.User{
		Email:        "example@mail.com",
		PasswordHash: "123",
		Version:      1,
		Role:         "student",
	}
	log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
)

func TestLogin(t *testing.T) {
	t.Run("Successful login", func(t *testing.T) {
		jwtService := &jwt.MockJWTService{}
		userStorage := &storage.MockUserStorage{}
		jwtStorage := &storage.MockJWTStorage{}
		blackListStorage := &storage.MockBlackListStorage{}

		auth := authpkg.NewAuthServiceImpl(
			jwtService,
			blackListStorage,
			jwtStorage,
			userStorage,
			log,
		)

		userForToken := model.User{
			Email:   correctUser.Email,
			Version: correctUser.Version,
			Role:    correctUser.Role,
		}

		userStorage.On("Login", correctUser.Email, correctUser.PasswordHash).Return(correctUser.Version, correctUser.Role, nil).Once()
		jwtService.On("CreateAccessToken", userForToken).Return("new_access_token", nil).Once()
		jwtService.On("CreateRefreshToken", userForToken).Return("new_refresh_token", nil).Once()
		jwtStorage.On("AddPair", "new_access_token", "new_refresh_token").Return(nil).Once()

		access, refresh, err := auth.Login(correctUser.Email, correctUser.PasswordHash)

		assert.NoError(t, err)
		assert.Equal(t, "new_access_token", access)
		assert.Equal(t, "new_refresh_token", refresh)

		jwtService.AssertExpectations(t)
		userStorage.AssertExpectations(t)
		jwtStorage.AssertExpectations(t)
	})

	t.Run("Failed login", func(t *testing.T) {
		jwtService := &jwt.MockJWTService{}
		userStorage := &storage.MockUserStorage{}
		jwtStorage := &storage.MockJWTStorage{}
		blackListStorage := &storage.MockBlackListStorage{}

		auth := authpkg.NewAuthServiceImpl(
			jwtService,
			blackListStorage,
			jwtStorage,
			userStorage,
			log,
		)

		userStorage.On("Login", "wrong@mail.com", "wrongpassword").Return(float64(-1), "", errors.New("incorrect email and password")).Once()

		access, refresh, err := auth.Login("wrong@mail.com", "wrongpassword")

		assert.Error(t, err)
		assert.Equal(t, "", access)
		assert.Equal(t, "", refresh)

		userStorage.AssertExpectations(t)
		jwtService.AssertNotCalled(t, "CreateAccessToken", "mock.Anything")
		jwtService.AssertNotCalled(t, "CreateRefreshToken", "mock.Anything")
		jwtStorage.AssertNotCalled(t, "AddPair", "mock.Anything", "mock.Anything")
	})
}

func TestRefresh(t *testing.T) {
	jwtService := &jwt.MockJWTService{}
	userStorage := &storage.MockUserStorage{}
	jwtStorage := &storage.MockJWTStorage{}
	blackListStorage := &storage.MockBlackListStorage{}

	auth := authpkg.NewAuthServiceImpl(
		jwtService,
		blackListStorage,
		jwtStorage,
		userStorage,
		log,
	)

	oldRefreshToken := "old_refresh_token"
	oldAccessToken := "old_access_token"
	user := model.User{Email: "test@test.com", Version: 1, Role: "user"}

	blackListStorage.On("IsAllowed", oldRefreshToken).Return(true, nil).Once()
	jwtService.On("GetUserInfo", oldRefreshToken).Return(user, nil).Once()
	userStorage.On("IsVersionValid", user.Email, user.Version).Return(true, nil).Once()
	jwtService.On("CreateAccessToken", user).Return("new_access_token", nil).Once()
	jwtService.On("CreateRefreshToken", user).Return("new_refresh_token", nil).Once()
	jwtStorage.On("GetAccessByRefresh", oldRefreshToken).Return(oldAccessToken, nil).Once()

	blackListStorage.On("AddTokens", []string{oldRefreshToken, oldAccessToken}).Return(nil).Once()
	blackListStorage.On("AddTokens", []string{oldRefreshToken}).Return(nil).Once()

	access, refresh, err := auth.Refresh(oldRefreshToken)

	assert.NoError(t, err)
	assert.Equal(t, "new_access_token", access)
	assert.Equal(t, "new_refresh_token", refresh)

	jwtService.AssertExpectations(t)
	userStorage.AssertExpectations(t)
	jwtStorage.AssertExpectations(t)
	blackListStorage.AssertExpectations(t)
}

func TestValidateToken(t *testing.T) {
	t.Run("Valid token", func(t *testing.T) {
		jwtService := &jwt.MockJWTService{}
		blackListStorage := &storage.MockBlackListStorage{}
		auth := authpkg.NewAuthServiceImpl(jwtService, blackListStorage, nil, nil, log)

		token := "valid_token"
		blackListStorage.On("IsAllowed", token).Return(true, nil).Once()
		jwtService.On("IsTokenValid", token).Return(true, nil).Once()

		isValid, err := auth.ValidateToken(token)

		assert.NoError(t, err)
		assert.True(t, isValid)
		blackListStorage.AssertExpectations(t)
		jwtService.AssertExpectations(t)
	})

	t.Run("Token in blacklist", func(t *testing.T) {
		jwtService := &jwt.MockJWTService{}
		blackListStorage := &storage.MockBlackListStorage{}
		auth := authpkg.NewAuthServiceImpl(jwtService, blackListStorage, nil, nil, log)

		token := "blacklisted_token"
		blackListStorage.On("IsAllowed", token).Return(false, nil).Once()

		isValid, err := auth.ValidateToken(token)

		assert.NoError(t, err)
		assert.False(t, isValid)
		blackListStorage.AssertExpectations(t)
		jwtService.AssertNotCalled(t, "IsTokenValid", "mock.Anything")
	})

	t.Run("Invalid token signature", func(t *testing.T) {
		jwtService := &jwt.MockJWTService{}
		blackListStorage := &storage.MockBlackListStorage{}
		auth := authpkg.NewAuthServiceImpl(jwtService, blackListStorage, nil, nil, log)

		token := "invalid_signature_token"
		blackListStorage.On("IsAllowed", token).Return(true, nil).Once()
		jwtService.On("IsTokenValid", token).Return(false, errors.New("bad signature")).Once()

		isValid, err := auth.ValidateToken(token)

		assert.Error(t, err)
		assert.False(t, isValid)
		blackListStorage.AssertExpectations(t)
		jwtService.AssertExpectations(t)
	})
}

func TestLogout(t *testing.T) {
	blackListStorage := &storage.MockBlackListStorage{}

	auth := authpkg.NewAuthServiceImpl(nil, blackListStorage, nil, nil, log)

	accessToken := "some_access_token"
	refreshToken := "some_refresh_token"

	blackListStorage.On("AddTokens", []string{accessToken, refreshToken}).Return(nil).Once()

	err := auth.Logout(accessToken, refreshToken)

	assert.NoError(t, err)
	blackListStorage.AssertExpectations(t)
}
