package pgsql

import (
	"fmt"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host string
	Port int
	User string
	Pass string
	Name string
}

type ConnectionPoolConfig struct {
	MaxIdleTime            time.Duration
	ConnectionMaxLifetime  time.Duration
	MaxIdleConnectionCount int
	MaxOpenConnectionCount int
	ConnectionTimeout      time.Duration
}

var (
	DefaultPoolConfig = &ConnectionPoolConfig{
		MaxIdleTime:            30 * time.Minute,
		ConnectionMaxLifetime:  2 * time.Hour,
		MaxIdleConnectionCount: 5,
		MaxOpenConnectionCount: 25,
		ConnectionTimeout:      30 * time.Second,
	}
	HighLoadPoolConfig = &ConnectionPoolConfig{
		MaxIdleTime:            15 * time.Minute,
		ConnectionMaxLifetime:  1 * time.Hour,
		MaxIdleConnectionCount: 10,
		MaxOpenConnectionCount: 50,
		ConnectionTimeout:      15 * time.Second,
	}
	LightPoolConfig = &ConnectionPoolConfig{
		MaxIdleTime:            1 * time.Hour,
		ConnectionMaxLifetime:  4 * time.Hour,
		MaxIdleConnectionCount: 3,
		MaxOpenConnectionCount: 10,
		ConnectionTimeout:      30 * time.Second,
	}
)

type Options struct {
	PoolConfig    *ConnectionPoolConfig
	SSLMode       string
	CommitTimeout time.Duration
}

type OptionFunc func(*Options)

func WithPoolConfig(c *ConnectionPoolConfig) OptionFunc {
	return func(o *Options) { o.PoolConfig = c }
}
func WithHighLoadPool() OptionFunc {
	return func(o *Options) { o.PoolConfig = HighLoadPoolConfig }
}
func WithLightPool() OptionFunc {
	return func(o *Options) { o.PoolConfig = LightPoolConfig }
}
func WithSSLMode(mode string) OptionFunc {
	return func(o *Options) { o.SSLMode = mode }
}
func WithCommitTimeout(d time.Duration) OptionFunc {
	return func(o *Options) { o.CommitTimeout = d }
}

func defaultOptions() *Options {
	return &Options{
		PoolConfig:    DefaultPoolConfig,
		SSLMode:       "disable",
		CommitTimeout: 10 * time.Second,
	}
}

func New(cfg *Config, optFuncs ...OptionFunc) (*gorm.DB, error) {
	opts := defaultOptions()
	for _, fn := range optFuncs {
		fn(opts)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.Name, opts.SSLMode,
		int(opts.PoolConfig.ConnectionTimeout.Seconds()),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := db.Use(otelgorm.NewPlugin(otelgorm.WithDBName(cfg.Name))); err != nil {
		return nil, fmt.Errorf("install otelgorm plugin: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying *sql.DB: %w", err)
	}
	sqlDB.SetConnMaxIdleTime(opts.PoolConfig.MaxIdleTime)
	sqlDB.SetConnMaxLifetime(opts.PoolConfig.ConnectionMaxLifetime)
	sqlDB.SetMaxIdleConns(opts.PoolConfig.MaxIdleConnectionCount)
	sqlDB.SetMaxOpenConns(opts.PoolConfig.MaxOpenConnectionCount)

	return db, nil
}
