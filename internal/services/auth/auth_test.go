package auth_test

import (
	"auth-service/internal/domain/models"
	"auth-service/internal/services/jwt"
	"time"

	authpkg "auth-service/internal/services/auth"

	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	correctUser = models.User{
		Email:        "example@mail.com",
		PasswordHash: []byte("123"),
		Version:      1,
		Role:         "student",
	}
	incorrectUser = models.User{
		Email:        "example@mail.com",
		PasswordHash: []byte("1234"),
		Version:      1,
		Role:         "student",
	}
)

func GetAuthService() authpkg.AuthServiceImpl {
	userStorage := mocUserStorage{
		users: []models.User{
			correctUser,
		},
	}
	jwtService := jwt.JWTServiceImpl{
		SecretKey:  []byte("a-string-secret-at-least-256-bits-long"),
		AccessTTL:  time.Duration(time.Minute * 15),
		RefreshTTL: time.Duration(time.Hour * 24),
	}
	return authpkg.AuthServiceImpl{
		BlackListStorage: &mockBlackListStorage{},
		JWTStorage:       &mockJWTStorage{},
		UserStorage:      &userStorage,
		JWTService:       &jwtService,
	}
}

// Moc BlackListStorage
type mockBlackListStorage struct {
	slice []string
}

func (s *mockBlackListStorage) AddTokens(tokens ...string) error {
	s.slice = append(s.slice, tokens...)
	return nil
}

func (s *mockBlackListStorage) IsAllowed(token string) (bool, error) {
	return !slices.Contains(s.slice, token), nil
}

// Moc UserStorage
type mocUserStorage struct {
	users []models.User
}

func (s *mocUserStorage) Login(email string, passwordHash []byte) (float64, string, error) {
	index := slices.IndexFunc(s.users,
		func(u models.User) bool {
			return u.Email == email && slices.Compare(u.PasswordHash, passwordHash) == 0
		},
	)

	if index == -1 {
		return -1, "", nil
	}

	return s.users[index].Version, s.users[index].Role, nil
}

func (s *mocUserStorage) IsVersionValid(email string, version float64) (bool, error) {
	return slices.ContainsFunc(s.users,
		func(u models.User) bool {
			return u.Version == version
		},
	), nil
}

type mockJWTStorage map[string]string

func (s *mockJWTStorage) AddPair(access string, refresh string) error {

	(*s)[refresh] = access
	return nil
}

func (s *mockJWTStorage) GetAccessByRefresh(refresh string) (string, error) {
	res := (*s)[refresh]
	delete((*s), refresh)

	return res, nil
}

// Tests

func TestLogin(t *testing.T) {
	auth := GetAuthService()

	access, refresh, err := auth.Login(correctUser.Email, correctUser.PasswordHash)
	assert.Nil(t, err)
	assert.NotNil(t, access, "invalid access token")
	assert.NotNil(t, refresh, "invalid refresh token")

	access, refresh, err = auth.Login(correctUser.Email, incorrectUser.PasswordHash)
	assert.NotNil(t, err)
	assert.Equal(t, "", access)
	assert.Equal(t, "", refresh)
}

func TestRefresh(t *testing.T) {
	auth := GetAuthService()

	_, refreshToken, err := auth.Login(correctUser.Email, correctUser.PasswordHash)
	assert.Nil(t, err)

	access, refresh, err := auth.Refresh(refreshToken)
	assert.Nil(t, err)
	assert.NotNil(t, access, "invalid access token")
	assert.NotNil(t, refresh, "invalid refresh token")
}

func TestValidateToken(t *testing.T) {
	auth := GetAuthService()

	access, refresh, err := auth.Login(correctUser.Email, correctUser.PasswordHash)
	assert.Nil(t, err)

	res, err := auth.ValidateToken(access)
	assert.Nil(t, err)
	assert.Equal(t, true, res)

	res, err = auth.ValidateToken(refresh)
	assert.Nil(t, err)
	assert.Equal(t, true, res)

	// TODO: реализовать более надёжный способ проверки поддельного JWT
	incurrectToken := access[:40] + "H" + access[41:]

	res, err = auth.ValidateToken(incurrectToken)
	assert.NotNil(t, err)
	assert.Equal(t, false, res)
}

func TestLogout(t *testing.T) {
	auth := GetAuthService()

	access, refresh, _ := auth.Login(correctUser.Email, correctUser.PasswordHash)

	err := auth.Logout(access, refresh)
	assert.Nil(t, err)

	res, err := auth.ValidateToken(access)
	assert.Nil(t, err)
	assert.Equal(t, false, res)
}
