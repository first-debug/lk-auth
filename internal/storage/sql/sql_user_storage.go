package sql

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"lk-auth/internal/libs/hash"
	sl "lk-auth/internal/libs/logger"
	"lk-auth/internal/storage"

	"gorm.io/gorm"
)

type SQLUserStorage struct {
	ctx    context.Context
	log    *slog.Logger
	client *gorm.DB
}

func NewSQLUserStorage(ctx context.Context, log *slog.Logger, wg *sync.WaitGroup, dbConfig gorm.Dialector, gormConfig *gorm.Config, pingTime time.Duration) (storage.UserStorage, error) {
	log.Info("Trying to connect to the database...")

	db, err := gorm.Open(dbConfig, gormConfig)
	if err != nil {
		return nil, err
	}

	// https://gorm.io/docs/dbresolver.html работа с множеством баз данных
	// https://gorm.io/docs/prometheus.html инеграция с Prometheus

	wg.Add(1)
	go func() {
		ticker := time.NewTicker(pingTime)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Debug("SQLUserStorage ping goroutine stopped")
				wg.Done()
				return
			case <-ticker.C:
				dbPinger, err := db.DB()
				if err != nil {
					log.Error("Fail to ping database", sl.Err(err))
					continue
				}
				if err := dbPinger.Ping(); err != nil {
					log.Error("SQLUserStorage didn't answer", "name", dbConfig.Name())
				}
			}
		}
	}()

	log.Info("Start auto migration...")

	err = db.AutoMigrate(User{})
	if err != nil {
		return nil, err
	}

	return &SQLUserStorage{
		ctx:    ctx,
		log:    log,
		client: db,
	}, nil
}

func (s *SQLUserStorage) Login(email, password string) (float64, string, error) {
	if email == "" || len(password) == 0 {
		s.log.Error("invalid input parameters")
		return -1, "", errors.New("invalid input parameters")
	}

	userInfo := User{}
	err := s.client.Take(&userInfo, "email = ?", email).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return -1, "", errors.New("user not found")
		}
		s.log.Error("database error", sl.Err(err))
		return -1, "", err
	}

	if len(userInfo.PasswordHash) == 0 || !hash.CheckPasswordHash([]byte(password), []byte(userInfo.PasswordHash)) {
		return -1, "", errors.New("incorrect email and password")
	}

	return userInfo.Version, userInfo.Role, nil
}

func (s *SQLUserStorage) IsVersionValid(email string, version float64) (bool, error) {
	if len(email) == 0 {
		return false, errors.New("email cannot be empty")
	}

	userInfo := User{}
	err := s.client.Take(&userInfo, "email = ?", email).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, errors.New("user not found")
		}
		return false, err
	}

	return userInfo.Version == version, nil
}

func (s *SQLUserStorage) ShutDown(context.Context) error {
	db, err := s.client.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
