package utils

import (
	"log/slog"
	"os"
	"strings"
	"time"
)

const DEFAULT_POLLING_INTERVAL = 30 * time.Second

type Config struct {
	APIEndpoint     string
	PollingInterval time.Duration
}

func LoadConfig(logger *slog.Logger) (Config, error) {
	var interval time.Duration

	if value := strings.TrimSpace(os.Getenv("POLLING_INTERVAL")); value != "" {
		var err error
		interval, err = time.ParseDuration(value)
		if err != nil {
			logger.Error(
				"invalid polling interval",
				"value", value,
				"error", err,
			)
			logger.Info(
				"using default polling interval",
				"interval", DEFAULT_POLLING_INTERVAL,
			)

			interval = DEFAULT_POLLING_INTERVAL
		}
	} else {
		logger.Info("using default polling interval", "interval", DEFAULT_POLLING_INTERVAL)
		interval = DEFAULT_POLLING_INTERVAL
	}

	return Config{
		APIEndpoint:     envOr("API_ENDPOINT", "127.0.0.1:8080"),
		PollingInterval: interval,
	}, nil
}

func envOr(name, defaultVal string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}

	return defaultVal
}
