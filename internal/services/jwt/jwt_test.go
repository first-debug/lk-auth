package jwt_test

import (
	"lk-auth/internal/domain/models"
	jwtpkg "lk-auth/internal/services/jwt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	jwtService jwtpkg.JWTService
	err        error
	user       = models.User{
		Email:        "example@mail.com",
		PasswordHash: "123",
		Version:      1,
		Role:         "student",
	}
	createFunc func(user models.User) (string, error)
)

func TestMain(t *testing.T) {
	jwtService, err = jwtpkg.NewJWTServiceImpl(
		[]byte("a-string-secret-at-least-256-bits-long"),
		time.Duration(time.Minute*15),
		time.Duration(time.Hour*24),
		slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			}),
		),
	)
	if err != nil {
		t.Fatal(err)
		t.Failed()
	}
	t.Run("CreateAccessToken", createAccessToken)
	t.Run("CreateRefreshToken", createRefreshToken)
}

func createAccessToken(t *testing.T) {
	createFunc = jwtService.CreateAccessToken
	t.Run("GetClaim", getClaim)
	t.Run("GetVersion", getVersion)
	t.Run("IsTokenValid", isTokenValid)
}

func createRefreshToken(t *testing.T) {
	createFunc = jwtService.CreateRefreshToken
	t.Run("GetClaim", getClaim)
	t.Run("GetVersion", getVersion)
	t.Run("IsTokenValid", isTokenValid)
}

func getClaim(t *testing.T) {
	actualToken, _ := createFunc(user)

	actualEmail, err := jwtService.GetEmail(actualToken)
	assert.Nil(t, err)
	assert.Equal(t, user.Email, actualEmail)

	actualVersion, err := jwtService.GetVersion(actualToken)
	assert.Nil(t, err)
	assert.Equal(t, user.Version, actualVersion)
}

func getVersion(t *testing.T) {
	token, err := createFunc(user)

	assert.Nil(t, err)

	currentVersion, err := jwtService.GetVersion(token)

	assert.Nil(t, err)

	assert.Equal(t, 1., currentVersion)
}

func isTokenValid(t *testing.T) {

	token, err := createFunc(user)

	assert.Nil(t, err)

	res, err := jwtService.IsTokenValid(token)

	assert.Nil(t, err)
	assert.Equal(t, true, res)

	builder := strings.Builder{}

	builder.WriteString(token[:len(token)-2])
	builder.WriteRune('J')
	res, err = jwtService.IsTokenValid(builder.String())

	assert.NotNil(t, err)
	assert.Equal(t, false, res)
	builder.Reset()

	builder.WriteString(token[:49])
	builder.WriteRune('J')
	builder.WriteString(token[48:])
	res, err = jwtService.IsTokenValid(builder.String())

	assert.NotNil(t, err)
	assert.Equal(t, false, res)
}
