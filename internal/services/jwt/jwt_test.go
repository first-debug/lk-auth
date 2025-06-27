package jwt_test

import (
	"auth-service/internal/domain/models"
	jwtpkg "auth-service/internal/services/jwt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	jwtService = jwtpkg.JWTServiceImpl{
		SecretKey: []byte("a-string-secret-at-least-256-bits-long"),
	}
	user = models.User{
		Email:        "example@mail.com",
		PasswordHash: []byte("123"),
		Version:      1,
	}
)

func TestGetClaim(t *testing.T) {
	actualToken, _ := jwtService.CreateToken(user, time.Duration(time.Minute*15))

	actualEmail, err := jwtService.GetEmail(actualToken)
	assert.Nil(t, err)
	assert.Equal(t, user.Email, actualEmail)

	actualVersion, err := jwtService.GetVersion(actualToken)
	assert.Nil(t, err)
	assert.Equal(t, user.Version, actualVersion)
}

func TestGetVersion(t *testing.T) {
	token, err := jwtService.CreateToken(user, time.Duration(time.Minute))

	assert.Nil(t, err)

	currentVersion, err := jwtService.GetVersion(token)

	assert.Nil(t, err)

	assert.Equal(t, 1., currentVersion)
}

func TestIsTokenValid(t *testing.T) {
	token, err := jwtService.CreateToken(user, time.Duration(time.Minute))

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
