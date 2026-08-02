package utils

import (
	"os"
	"strings"
)

type Config struct {
	APIEndpoint string
}

func LoadConfig() (Config, error) {
	return Config{
		APIEndpoint: envOr("API_ENDPOINT", "127.0.0.1:8080"),
	}, nil
}

func envOr(name, defaultVal string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}

	return defaultVal
}
