package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	Environment  string
	DatabaseURL  string
	RedisURL     string
	KafkaBrokers []string
	JWTSecret    string
	ServiceName  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	return &Config{
		Port:         getEnv("PORT", "3002"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		DatabaseURL:  getEnv("DATABASE_URL", "root:password@tcp(localhost:3306)/quckapp_workspaces?parseTime=true"),
		RedisURL:     getEnv("REDIS_URL", "localhost:6379"),
		KafkaBrokers: strings.Split(kafkaBrokers, ","),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key"),
		ServiceName:  "workspace-service",
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
