package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

type Config struct {
	Port                string
	StorageBucket       string
	AllowedOrigin       string
	FirebasePrivateKey  string
	FirebaseClientEmail string
	ProjectID           string
	Opt                 option.ClientOption
}

func LoadConfig() *Config {
	// Load environment-specific .env file
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}

	envFile := ".env." + env
	if err := godotenv.Load(envFile); err != nil {
		println("Warning: Error loading " + envFile)
	}

	// Get credentials path and convert to absolute if needed
	credentialsPath := os.Getenv("CREDENTIALS")
	if credentialsPath != "" && !filepath.IsAbs(credentialsPath) {
		workDir, _ := os.Getwd()
		credentialsPath = filepath.Join(workDir, credentialsPath)
	}

	// Add debug logging
	log.Printf("Loading credentials from: %s", credentialsPath)
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		log.Printf("Warning: Credentials file not found at %s", credentialsPath)
	}

	return &Config{
		Port:                getEnv("PORT", "8080"),
		StorageBucket:       os.Getenv("FIREBASE_STORAGE_BUCKET"),
		AllowedOrigin:       os.Getenv("ALLOWED_ORIGIN"),
		FirebasePrivateKey:  os.Getenv("FIREBASE_PRIVATE_KEY"),
		FirebaseClientEmail: os.Getenv("FIREBASE_CLIENT_EMAIL"),
		ProjectID:           os.Getenv("FIREBASE_PROJECT_ID"),
		Opt:                 option.WithCredentialsFile(credentialsPath),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
