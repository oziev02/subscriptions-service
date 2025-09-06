package config

import (
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv   string
	LogLevel string
	HTTP     HTTPConfig
	DB       DBConfig
}

type HTTPConfig struct {
	Port int
}

type DBConfig struct {
	DSN string
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AutomaticEnv()

	_ = v.ReadInConfig()

	cfg := &Config{
		AppEnv:   getEnv("APP_ENV", "dev"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		HTTP: HTTPConfig{
			Port: getEnvInt("HTTP_PORT", 8080),
		},
		DB: DBConfig{
			DSN: getEnv("DB_DSN", "postgres://postgres:postgres@localhost:5432/subscriptions?sslmode=disable"),
		},
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func getEnvInt(key string, def int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return def
}
