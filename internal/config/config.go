package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        string
	DBPath      string
	Latitude    float64
}

func Load() *Config {
	return &Config{
		Port:     getEnv("PORT", "8080"),
		DBPath:   getEnv("DB_PATH", "./ctd_data.db"),
		Latitude: getEnvFloat("LATITUDE", 30.0),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}
