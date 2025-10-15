package sql

import (
	"lk-auth/internal/config"
	"log/slog"
	"time"

	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitGorm(cfg *config.Config, log *slog.Logger) (gorm.Dialector, *gorm.Config) {
	dbOptions := mysql.New(
		mysql.Config{
			DSN: cfg.Storages.SQL,
		},
	)

	gormSlogOpt := append(
		[]slogGorm.Option{},
		slogGorm.WithHandler(log.Handler()),
		slogGorm.WithSlowThreshold(time.Millisecond*200),
		slogGorm.SetLogLevel(slogGorm.DefaultLogType, slog.LevelDebug),
	)

	if cfg.Env == "prod" {
		gormSlogOpt = append(
			gormSlogOpt,
			slogGorm.WithIgnoreTrace(),
		)
	} else if *cfg.Logger.Level == slog.LevelDebug {
		gormSlogOpt = append(
			gormSlogOpt,
			slogGorm.WithTraceAll(),
		)
	}
	gormCfg := &gorm.Config{
		Logger:               slogGorm.New(gormSlogOpt...),
		PrepareStmt:          true,
		DisableAutomaticPing: true,
	}

	return dbOptions, gormCfg
}
