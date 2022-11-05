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
