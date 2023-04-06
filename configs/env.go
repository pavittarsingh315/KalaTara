package configs

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func InitEnv() {
	err := godotenv.Load() // will load vars from .env file into ENV for current process
	if err != nil {
		log.Fatalf("Error initializing .env file: %v", err)
	}
}

// returns true if app is in production mode
func EnvProdActive() bool {
	value, exists := os.LookupEnv("APP_ENV")
	if !exists {
		log.Fatal("APP_ENV not set")
	}
	if value != "production" && value != "development" {
		log.Fatalf("APP_ENV must be either \"production\" or \"development\"")
	}
	return value == "production"
}

func EnvTokenSecrets() (access, refresh string) {
	value1, exists1 := os.LookupEnv("ACCESS_TOKEN_SECRET")
	if !exists1 {
		log.Fatal("ACCESS_TOKEN_SECRET not set")
	}
	value2, exists2 := os.LookupEnv("REFRESH_TOKEN_SECRET")
	if !exists2 {
		log.Fatal("REFRESH_TOKEN_SECRET not set")
	}
	access, refresh = value1, value2
	return
}

func EnvSendGridKeyAndFrom() (key, sender string) {
	value1, exists1 := os.LookupEnv("SENDGRID_API_KEY")
	if !exists1 {
		log.Fatal("SENDGRID_API_KEY not set")
	}
	value2, exists2 := os.LookupEnv("SENDGRID_SENDER")
	if !exists2 {
		log.Fatal("SENDGRID_SENDER not set")
	}
	key, sender = value1, value2
	return
}

func EnvTwilioIDKeyFrom() (id, token, from string) {
	value1, exists1 := os.LookupEnv("TWILIO_ACCOUNT_SID")
	if !exists1 {
		log.Fatal("TWILIO_ACCOUNT_SID not set")
	}
	value2, exists2 := os.LookupEnv("TWILIO_AUTH_TOKEN")
	if !exists2 {
		log.Fatal("TWILIO_AUTH_TOKEN not set")
	}
	value3, exists3 := os.LookupEnv("TWILIO_FROM_NUMBER")
	if !exists3 {
		log.Fatal("TWILIO_FROM_NUMBER not set")
	}
	id, token, from = value1, value2, value3
	return
}

func EnvRedisAddr() string {
	value, exists := os.LookupEnv("REDIS_ADDRESS")
	if !exists {
		log.Fatal("REDIS_ADDRESS not set")
	}
	return value
}

func EnvRedisUsername() string {
	value, exists := os.LookupEnv("REDIS_USERNAME")
	if !exists {
		log.Fatal("REDIS_USERNAME not set")
	}
	return value
}

func EnvRedisPassword() string {
	value, exists := os.LookupEnv("REDIS_PASSWORD")
	if !exists {
		log.Fatal("REDIS_PASSWORD not set")
	}
	return value
}

func EnvRedisDatabase() string {
	value, exists := os.LookupEnv("REDIS_DATABASE")
	if !exists {
		log.Fatal("REDIS_DATABASE not set")
	}
	return value
}

func EnvPostgresDNS() string {
	value, exists := os.LookupEnv("POSTGRES_DNS")
	if !exists {
		log.Fatal("POSTGRES_DNS not set")
	}
	return value
}

func EnvDbMaxOpenConns() int {
	value, exists := os.LookupEnv("DB_MAX_OPEN_CONNS")
	if !exists {
		log.Fatal("DB_MAX_OPEN_CONNS not set")
	}
	max, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("Error converting DB_MAX_OPEN_CONNS to int: %v", err)
	}
	return max
}

func EnvDbMaxIdleConns() int {
	value, exists := os.LookupEnv("DB_MAX_IDLE_CONNS")
	if !exists {
		log.Fatal("DB_MAX_IDLE_CONNS not set")
	}
	max, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("Error converting DB_MAX_IDLE_CONNS to int: %v", err)
	}
	return max
}

func EnvDbConnMaxLifetime() int {
	value, exists := os.LookupEnv("DB_MAX_CONN_LIFETIME")
	if !exists {
		log.Fatal("DB_MAX_CONN_LIFETIME not set")
	}
	max, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("Error converting DB_MAX_CONN_LIFETIME to int: %v", err)
	}
	return max
}
