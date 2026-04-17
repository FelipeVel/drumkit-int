package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all environment-driven configuration for the service.
type Config struct {
	ServerPort          string
	TurvoBaseURL        string
	TurvoAPIKey         string
	TurvoClientID       string
	TurvoClientSecret   string
	TurvoUsername       string
	TurvoPassword       string
	HTTPTimeoutSec      int
	CORSAllowedOrigins  string
}

// Load reads configuration from a .env file (if present) and then from
// environment variables, applying sensible defaults where needed.
func Load() *Config {
	// Best-effort: load .env if it exists; ignore the error if it doesn't.
	_ = godotenv.Load()

	return &Config{
		ServerPort:         getEnv("SERVER_PORT", ":8080"),
		TurvoBaseURL:       getEnv("TURVO_BASE_URL", ""),
		TurvoAPIKey:        getEnv("TURVO_API_KEY", ""),
		TurvoClientID:      getEnv("TURVO_CLIENT_ID", ""),
		TurvoClientSecret:  getEnv("TURVO_CLIENT_SECRET", ""),
		TurvoUsername:      getEnv("TURVO_USERNAME", ""),
		TurvoPassword:      getEnv("TURVO_PASSWORD", ""),
		HTTPTimeoutSec:     getEnvInt("HTTP_TIMEOUT_SEC", 10),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return defaultVal
}
