package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppPort string
	DB      DBConfig
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
}

func Load() (Config, error) {
	dbPort, err := getEnvAsInt("DB_PORT", 5432)

	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	cfg := Config{
		AppPort: getEnv("APP_PORT", "7000"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "todo_api"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "UTC"),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) (int, error) {
	if value, ok := os.LookupEnv(key); ok {
		n, err := strconv.Atoi(value)
		if err != nil {
			return 0, err
		}
		return n, nil
	}
	return defaultValue, nil
}
