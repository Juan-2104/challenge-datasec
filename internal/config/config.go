package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
    Server     ServerConfig
    MetadataDB MetadataDBConfig
    Security   SecurityConfig
    Logging    LoggingConfig
    API        APIConfig
}

type ServerConfig struct {
	Port    int
	GinMode string
}

type MetadataDBConfig struct {
    Host     string
    Port     int
    Username string
    Password string
    Database string
    Params   string
}

type SecurityConfig struct {
	EncryptionKey string
	JWTSecret     string
}

type LoggingConfig struct {
	Level  string
	Format string
}

type APIConfig struct {
	Version string
	Timeout time.Duration
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using environment variables")
	}

    cfg := &Config{
        Server: ServerConfig{
            Port:    getIntEnv("PORT", 8080),
            GinMode: getStringEnv("GIN_MODE", "release"),
        },
        MetadataDB: MetadataDBConfig{
            Host:     getStringEnv("METADATA_DB_HOST", "localhost"),
            Port:     getIntEnv("METADATA_DB_PORT", 3306),
            Username: getStringEnv("METADATA_DB_USER", "metauser"),
            Password: getStringEnv("METADATA_DB_PASSWORD", "metapass"),
            Database: getStringEnv("METADATA_DB_NAME", "classifier_meta"),
            Params:   getStringEnv("METADATA_DB_PARAMS", "parseTime=true&charset=utf8mb4&loc=UTC"),
        },
        Security: SecurityConfig{
            EncryptionKey: getStringEnv("ENCRYPTION_KEY", ""),
            JWTSecret:     getStringEnv("JWT_SECRET", ""),
        },
        Logging: LoggingConfig{
            Level:  getStringEnv("LOG_LEVEL", "info"),
            Format: getStringEnv("LOG_FORMAT", "json"),
        },
        API: APIConfig{
            Version: getStringEnv("API_VERSION", "v1"),
            Timeout: getDurationEnv("API_TIMEOUT", 30*time.Second),
        },
    }

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
    if c.Security.EncryptionKey == "" {
        return fmt.Errorf("ENCRYPTION_KEY is required")
    }
    if len(c.Security.EncryptionKey) != 32 {
        return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 characters")
    }
    if c.Security.JWTSecret == "" {
        return fmt.Errorf("JWT_SECRET is required")
    }
    if c.MetadataDB.Host == "" {
        return fmt.Errorf("METADATA_DB_HOST is required")
    }
    if c.MetadataDB.Username == "" {
        return fmt.Errorf("METADATA_DB_USER is required")
    }
    if c.MetadataDB.Password == "" {
        return fmt.Errorf("METADATA_DB_PASSWORD is required")
    }
    if c.MetadataDB.Database == "" {
        return fmt.Errorf("METADATA_DB_NAME is required")
    }
    return nil
}

func getStringEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
