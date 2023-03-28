package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func InitEnv() {
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load() // will load vars from .env file into ENV for current process
		if err != nil {
			log.Fatal("Error initializing .env file")
		}
	}
}

func EnvTokenSecrets() (access, refresh string) {
	access = os.Getenv("ACCESS_TOKEN_SECRET")
	refresh = os.Getenv("REFRESH_TOKEN_SECRET")
	return
}

func EnvSendGridKeyAndFrom() (key, sender string) {
	key = os.Getenv("SENDGRID_API_KEY")
	sender = os.Getenv("SENDGRID_SENDER")
	return
}

func EnvTwilioIDKeyFrom() (id, token, from string) {
	id = os.Getenv("TWILIO_ACCOUNT_SID")
	token = os.Getenv("TWILIO_AUTH_TOKEN")
	from = os.Getenv("TWILIO_FROM_NUMBER")
	return
}

func EnvPostgresDNS() (dns string) {
	dns = os.Getenv("POSTGRES_DNS")
	return
}

func EnvRedisAddr() (addr string) {
	addr = os.Getenv("REDIS_ADDRESS")
	return
}

func EnvRedisUsername() (username string) {
	username = os.Getenv("REDIS_USERNAME")
	return
}

func EnvRedisPassword() (password string) {
	password = os.Getenv("REDIS_PASSWORD")
	return
}

func EnvRedisDatabase() (database string) {
	database = os.Getenv("REDIS_DATABASE")
	return
}
