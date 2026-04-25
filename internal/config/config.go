package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {

	// Server
	Port         string
	ReadTimeout  int
	WriteTimeout int

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Auth
	JWTSecret     string
	JWTExpiration int

	// Service
	ServiceName string

	// DBRetry
	DBPingTimeout      int
	DBPingAttempts     int
	DBPingDelaySeconds int

	ShutdownTimeout int
}

// Now load all the secrets from env variables

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	var missing []string

	cfg.Port = getEnvOrDefault("PORT", "8000")
	cfg.ReadTimeout = getEnvIntOrDefault("READ_TIMEOUT_SECONDS", 10)
	cfg.WriteTimeout = getEnvIntOrDefault("WRITE_TIMEOUT_SECONDS", 10)

	cfg.DBHost = requireEnv("DB_HOST", &missing)
	cfg.DBPort = getEnvOrDefault("DB_PORT", "5432")
	cfg.DBUser = requireEnv("DB_USER", &missing)
	cfg.DBPassword = requireEnv("DB_PASSWORD", &missing)
	cfg.DBName = requireEnv("DB_Name", &missing)
	cfg.DBSSLMode = getEnvOrDefault("DB_SSL_MODE", "disabled")

	cfg.JWTExpiration = getEnvIntOrDefault("JWT_EXPIRATION_HOURS", 24)
	cfg.JWTSecret = requireEnv("JWT_SECRET", &missing)

	cfg.ServiceName = getEnvOrDefault("SERVICE_NAME", "user_service")

	cfg.DBPingTimeout = getEnvIntOrDefault("DB_PING_TIMEOUT", 5)
	cfg.DBPingAttempts = getEnvIntOrDefault("DB_PING_ATTEMPTS", 3)
	cfg.DBPingDelaySeconds = getEnvIntOrDefault("DB_DELAY_SECONDS", 2)

	cfg.ShutdownTimeout = getEnvIntOrDefault("SHUT_DOWN_TIMEOUT", 30)

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required env variables %v", missing)
	}
	return cfg, nil
}

func getEnvOrDefault(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvIntOrDefault(key string, defaultInt int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultInt
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultInt
	}
	return n
}

func requireEnv(key string, missing *[]string) string {
	val := os.Getenv(key)
	if val == "" {
		*missing = append(*missing, key)
	}
	return val
}

func (cfg *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)
}
