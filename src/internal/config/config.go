package config

import (
	"os"
)

// TODO(pencelheimer): maybe we should split this into domain-specific sub-structs (App, Auth0, DB)?
type Config struct {
	AppPort           string
	AppEnv            string
	AppBaseUrl        string

	Auth0Domain       string
	Auth0ClientID     string
	Auth0ClientSecret string
	Auth0CallbackURL  string

	DbHost            string
	DbUser            string
	DbPassword        string
	DbName            string
	DbPort            string
	DbSSLMode         string
}

func Load() *Config {
	return &Config{
		AppPort:    os.Getenv("APP_PORT"),
		AppEnv:     os.Getenv("APP_ENV"),
		AppBaseUrl: os.Getenv("APP_BASE_URL"),

		Auth0Domain:       os.Getenv("AUTH0_DOMAIN"),
		Auth0ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
		Auth0ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		Auth0CallbackURL:  os.Getenv("AUTH0_CALLBACK_URL"),

		DbHost:     os.Getenv("DB_HOST"),
		DbUser:     os.Getenv("DB_USER"),
		DbPassword: os.Getenv("DB_PASSWORD"),
		DbName:     os.Getenv("DB_NAME"),
		DbPort:     os.Getenv("DB_PORT"),
		DbSSLMode:  os.Getenv("DB_SSL_MODE"),
	}
}
