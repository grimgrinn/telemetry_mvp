package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort  string
	DBConn      string
	LogLevel    string
	KafkaBroker string
}

// Load загружает конфигурацию из переменных окружения
// если переменная не задана - использует значение по умолчанию
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using system env")
	}

	cfg := &Config{
		ServerPort:  getEnv("SERVER_PORT", "50051"),
		DBConn:      getEnv("DB_CONN", ""),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		KafkaBroker: getEnv("KAFKA_BROKER", "localhost:9092"),
	}

	if cfg.DBConn == "" {
		log.Fatal("DB_CONN environment variable is required")

	}
	return cfg
}

// getEnv возвращает значение переменной окружения или default, если не задана
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

// GetEnvAsInt возвращает int из переменной окружения или default
func GetEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
