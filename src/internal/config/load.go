package config

import "os"

func Load() *Config {
	return &Config{
		AppAddress:        os.Getenv("APP_ADDRESS"),
		AppPort:           os.Getenv("APP_PORT"),
		AppEnv:            os.Getenv("APP_ENV"),
		AppBaseUrl:        os.Getenv("APP_BASE_URL"),
		SessionSecret:     os.Getenv("SESSION_SECRET"),
		Auth0Domain:       os.Getenv("AUTH0_DOMAIN"),
		Auth0ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
		Auth0ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		Auth0CallbackURL:  os.Getenv("AUTH0_CALLBACK_URL"),
		DbHost:            os.Getenv("DB_HOST"),
		DbUser:            os.Getenv("DB_USER"),
		DbPassword:        os.Getenv("DB_PASSWORD"),
		DbName:            os.Getenv("DB_NAME"),
		DbPort:            os.Getenv("DB_PORT"),
		DbSSLMode:         os.Getenv("DB_SSL_MODE"),
	}
}
