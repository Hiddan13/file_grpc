package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	StoragePath   string
	UploadLimit   int
	DownloadLimit int
	ListLimit     int
	GRPCPort      string
}

func Load() *Config {
	// Загружаем .env файл, если он существует
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}
	return &Config{
		StoragePath:   getEnv("STORAGE_PATH", "./uploads_default"),
		UploadLimit:   getEnvAsInt("UPLOAD_LIMIT", 10),
		DownloadLimit: getEnvAsInt("DOWNLOAD_LIMIT", 10),
		ListLimit:     getEnvAsInt("LIST_LIMIT", 100),
		GRPCPort:      getEnv("GRPC_PORT", ":50051"),
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}
