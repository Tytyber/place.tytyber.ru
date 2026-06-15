package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DatabaseHost     string
	DatabasePort     int
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string

	// Server
	ServerHost string
	ServerPort int

	// Session
	SessionSecretKey string

	// Application
	AppName  string
	AppURL   string
	DarkMode bool
}

var config *Config

func Load() *Config {
	if config != nil {
		return config
	}

	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	config = &Config{
		// Database
		DatabaseHost:     getEnv("DATABASE_HOST", "localhost"),
		DatabasePort:     getEnvAsInt("DATABASE_PORT", 5432),
		DatabaseUser:     getEnv("DATABASE_USER", "postgres"),
		DatabasePassword: getEnv("DATABASE_PASSWORD", ""),
		DatabaseName:     getEnv("DATABASE_NAME", "tytyber_club"),

		// Server
		ServerHost: getEnv("SERVER_HOST", "localhost"),
		ServerPort: getEnvAsInt("SERVER_PORT", 8080),

		// Session
		SessionSecretKey: getEnv("SESSION_SECRET_KEY", "place-tytyber-secret-key-change-in-production"),

		// Application
		AppName:  getEnv("APP_NAME", "Tytyber Place"),
		AppURL:   getEnv("APP_URL", "http://localhost:8080"),
		DarkMode: getEnvAsBool("DARK_MODE_ENABLED", false),
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("Database: %s@%s:%d/%s", config.DatabaseUser, config.DatabaseHost, config.DatabasePort, config.DatabaseName)
	log.Printf("Server: %s:%d", config.ServerHost, config.ServerPort)
	log.Printf("Dark Mode: %v", config.DarkMode)

	return config
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

func Get() *Config {
	if config == nil {
		return Load()
	}
	return config
}
