package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/tkanos/gonfig"
)

type Configuration struct {
	Port                   string `env:"PORT"`
	Dburl                  string `env:"DB_URL"`
	DbHost                 string `env:"DB_HOST"`
	DbPort                 string `env:"DB_PORT"`
	DbName                 string `env:"DB_NAME"`
	DbUser                 string `env:"DB_USER"`
	DbPassword             string `env:"DB_PASSWORD"`
	JWTSecret              string `env:"JWT_SECRET"`
	PrimaryEmail           string `env:"PRIMARY_EMAIL"`
	PrimaryEmailPassword   string `env:"PRIMARY_EMAIL_PASSWORD"`
	FrontendUrl            string `env:"FRONTEND_URL"`
	GoogleClientID         string `env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret     string `env:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL      string `env:"GOOGLE_REDIRECT_URL"`
	GithubAuthClientID     string `env:"GITHUB_AUTH_CLIENT_ID"`
	GithubAuthClientSecret string `env:"GITHUB_AUTH_CLIENT_SECRET"`
	GithubAuthRedirectURL  string `env:"GITHUB_AUTH_REDIRECT_URL"`
}

func GetConfig() Configuration {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("error loading dotenv file", err)
	}
	configuration := Configuration{}
	err = gonfig.GetConf("", &configuration)
	if err != nil {
		fmt.Println("error in config:", err)
	}
	return configuration
}
