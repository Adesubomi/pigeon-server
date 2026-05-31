package db

import (
	"context"
	"database/sql"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/adesubomi/pigeon-server/config"
	"github.com/adesubomi/pigeon-server/internal/domain/auth"
	"github.com/adesubomi/pigeon-server/internal/domain/device"
	"github.com/adesubomi/pigeon-server/internal/domain/endpoint"
	"github.com/adesubomi/pigeon-server/internal/domain/event"
	"github.com/adesubomi/pigeon-server/internal/domain/push"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	Gorm  *gorm.DB
	sqlDB *sql.DB
}

func Connect(ctx context.Context, cfg *config.Config) (*DB, error) {
	query := url.Values{}
	query.Set("sslmode", cfg.DBSSLMode)
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.DBUser, cfg.DBPassword),
		Host:     net.JoinHostPort(cfg.DBHost, strconv.Itoa(cfg.DBPort)),
		Path:     cfg.DBName,
		RawQuery: query.Encode(),
	}
	gormDB, err := gorm.Open(postgres.Open(dsn.String()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	if err := gormDB.AutoMigrate(
		&auth.User{},
		&endpoint.Endpoint{},
		&endpoint.PairingCode{},
		&device.Device{},
		&event.Event{},
		&push.DeliveryLog{},
	); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return &DB{Gorm: gormDB, sqlDB: sqlDB}, nil
}

func (d *DB) Close() error {
	if d == nil || d.sqlDB == nil {
		return nil
	}
	return d.sqlDB.Close()
}
