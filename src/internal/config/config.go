package config

import (
	"os"
)

type Config struct {
	AppPort           string
	AppEnv            string
	AppBaseUrl        string
	Auth0Domain       string
	Auth0ClientID     string
	Auth0ClientSecret string
	Auth0CallbackURL  string
}

func Load() *Config {
	return &Config{
		AppPort:           os.Getenv("APP_PORT"),
		AppEnv:            os.Getenv("APP_ENV"),
		AppBaseUrl:        os.Getenv("APP_BASE_URL"),
		Auth0Domain:       os.Getenv("AUTH0_DOMAIN"),
		Auth0ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
		Auth0ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		Auth0CallbackURL:  os.Getenv("AUTH0_CALLBACK_URL"),
	}
}
