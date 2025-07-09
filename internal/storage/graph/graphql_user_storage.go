package storage

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"lk-auth/internal/generated/genqlient"
	"lk-auth/internal/libs/hash"
	sl "lk-auth/internal/libs/logger"
	"lk-auth/internal/storage"

	"github.com/Khan/genqlient/graphql"
)

type GraphQLUserStorage struct {
	ctx    context.Context
	client graphql.Client
	log    *slog.Logger
}

func NewGraphQLUserStorage(ctx context.Context, wg *sync.WaitGroup, log *slog.Logger, pingTime time.Duration, url string) storage.UserStorage {
	client := graphql.NewClient(url, http.DefaultClient)

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.Ticker{}
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// TODO: вызывать прверку доступности сервиса
			}
		}
	}()

	return &GraphQLUserStorage{
		ctx:    ctx,
		client: client,
		log:    log,
	}
}

func (s *GraphQLUserStorage) Login(email, password string) (dataVersion float64, role string, err error) {
	if email == "" || len(password) == 0 {
		s.log.Error("invalid input parameters")
		return -1, "", errors.New("invalid input parameters")
	}

	res, err := genqlient.GetUserInfo(s.ctx, s.client, email)
	if err != nil {
		s.log.Error("GraphQLUserStorage Login", sl.Err(err))
		return -1, "", errors.New("user not found")
	}
	if res.AuthUser.Email == "" {
		return -1, "", errors.New("user not found")
	}

	if len(res.AuthUser.PasswordHash) == 0 || !hash.CheckPasswordHash([]byte(password), []byte(res.AuthUser.PasswordHash)) {
		return -1, "", errors.New("incorrect email and password")
	}

	return float64(res.AuthUser.Version), res.AuthUser.Role, nil

}

func (s *GraphQLUserStorage) IsVersionValid(email string, version float64) (bool, error) {
	if len(email) == 0 {
		return false, errors.New("email cannot be empty")
	}

	res, err := genqlient.GetUserInfo(s.ctx, s.client, email)
	if err != nil {
		return false, errors.New("user not found")
	}

	return res.AuthUser.Version == int(version), nil
}

func (s *GraphQLUserStorage) ShutDown(context.Context) error {
	return nil
}
