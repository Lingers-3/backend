package config

//go:generate bash -c "awk -f gen_load.awk $GOFILE | gofmt > load.go"
type Config struct {
	AppAddress string
	AppPort    string
	AppEnv     string
	AppBaseUrl string

	SessionSecret string

	Auth0Domain       string
	Auth0ClientID     string
	Auth0ClientSecret string
	Auth0CallbackURL  string

	DbHost     string
	DbUser     string
	DbPassword string
	DbName     string
	DbPort     string
	DbSSLMode  string
}
