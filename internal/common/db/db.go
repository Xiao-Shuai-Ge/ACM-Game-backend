package db

import (
	"context"

	"acmgame-backend/internal/common/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB(cfg *config.Config, logger *zap.Logger) (*gorm.DB, error) {
	// If DSN is empty, we might want to skip or fail.
	// For now, let's assume it's required if we include this module.
	if cfg.MySQL.DSN == "" {
		logger.Warn("MySQL DSN is empty, skipping connection")
		return nil, nil
	}

	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect database", zap.Error(err))
		return nil, err
	}

	logger.Info("connected to mysql database")
	return db, nil
}

var Module = fx.Options(
	fx.Provide(NewDB),
	fx.Invoke(func(lc fx.Lifecycle, db *gorm.DB) {
		if db != nil {
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					sqlDB, _ := db.DB()
					if sqlDB != nil {
						return sqlDB.Close()
					}
					return nil
				},
			})
		}
	}),
)
