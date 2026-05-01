package pgsql

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBConfig struct {
	Host string
	Port int
	User string
	Pass string
	Name string
}

type Config struct {
	DB            *DBConfig
	MigrationsDir string
	Steps         int
}

type PgsqlMigrator struct {
	cfg *Config
	db  *sql.DB
}

func NewPgsqlMigrator(cfg *Config) (*PgsqlMigrator, error) {
	// negative is operator error; 0 means "use default" (coerced to 1); upper clamp guards against typos causing destructive runs
	if cfg.Steps < 0 {
		return nil, fmt.Errorf("invalid steps %d: must be >= 0", cfg.Steps)
	}
	if cfg.Steps == 0 {
		cfg.Steps = 1
	}
	if cfg.Steps > 1000 {
		return nil, fmt.Errorf("invalid steps %d: must be <= 1000", cfg.Steps)
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Pass, cfg.DB.Name)

	conn, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db, err := conn.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	return &PgsqlMigrator{cfg: cfg, db: db}, nil
}

func (pm *PgsqlMigrator) Up() error {
	log.Println("Migrating up...")
	if err := goose.Up(pm.db, pm.cfg.MigrationsDir); err != nil {
		return fmt.Errorf("goose up failed: %w", err)
	}
	return nil
}

func (pm *PgsqlMigrator) Down() error {
	log.Printf("Rolling back %d step(s)...\n", pm.cfg.Steps)
	for i := range pm.cfg.Steps {
		if err := goose.Down(pm.db, pm.cfg.MigrationsDir); err != nil {
			return fmt.Errorf("goose down step %d failed: %w", i+1, err)
		}
	}
	return nil
}

func (pm *PgsqlMigrator) DownAll() error {
	log.Println("Migrating down all...")
	if err := goose.DownTo(pm.db, pm.cfg.MigrationsDir, 0); err != nil {
		return fmt.Errorf("goose down all failed: %w", err)
	}
	return nil
}

func (pm *PgsqlMigrator) Status() error {
	log.Println("Checking migration status...")
	if err := goose.Status(pm.db, pm.cfg.MigrationsDir); err != nil {
		return fmt.Errorf("goose status failed: %w", err)
	}
	return nil
}

func (pm *PgsqlMigrator) Close() error {
	if pm.db != nil {
		return pm.db.Close()
	}
	return nil
}
