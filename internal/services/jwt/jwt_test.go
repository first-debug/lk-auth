package jwt_test

import (
	"auth-service/internal/domain/models"
	"auth-service/internal/services/jwt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	jwtService = jwt.JWTServiceImpl{
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
