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

func EnvMySqlDNS() (dns string) {
	dns = os.Getenv("MYSQL_DNS")
	return
}
