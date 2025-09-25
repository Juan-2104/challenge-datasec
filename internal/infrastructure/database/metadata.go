package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"database-classifier/internal/config"
)

const (
	metadataMaxOpenConns = 25
	metadataMaxIdleConns = 5
	metadataConnMaxLifetime = time.Hour
)

func NewMetadataDB(cfg *config.MetadataDBConfig) (*sql.DB, error) {
	dsn := buildDSN(cfg)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata database: %w", err)
	}

	db.SetMaxOpenConns(metadataMaxOpenConns)
	db.SetMaxIdleConns(metadataMaxIdleConns)
	db.SetConnMaxLifetime(metadataConnMaxLifetime)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping metadata database: %w", err)
	}

	return db, nil
}

func buildDSN(cfg *config.MetadataDBConfig) string {
	params := cfg.Params
	if params == "" {
		params = "parseTime=true&charset=utf8mb4&loc=UTC"
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		params,
	)
}

