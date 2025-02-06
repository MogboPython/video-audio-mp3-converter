package config

import (
	"os"

	"google.golang.org/api/option"
)

type Config struct {
	Port                string
	StorageBucket       string
	AllowedOrigin       string
	FirebasePrivateKey  string
	FirebaseClientEmail string
	Opt                 option.ClientOption
}

func LoadConfig() *Config {
	return &Config{
		Port:                getEnv("PORT", "8080"),
		StorageBucket:       os.Getenv("FIREBASE_STORAGE_BUCKET"),
		AllowedOrigin:       os.Getenv("ALLOWED_ORIGIN"),
		FirebasePrivateKey:  os.Getenv("FIREBASE_PRIVATE_KEY"),
		FirebaseClientEmail: os.Getenv("FIREBASE_CLIENT_EMAIL"),
		Opt:                 option.WithCredentialsFile(os.Getenv("CREDENTIALS")),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
