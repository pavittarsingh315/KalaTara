package configs

import (
	"log"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var (
	Redis *redis.Client
)

func InitRedis() {
	db, err := strconv.Atoi(EnvRedisDatabase())
	if err != nil {
		log.Fatal("Error connecting to Redis...")
		panic(err)
	}

	Redis = redis.NewClient(&redis.Options{
		Addr: EnvRedisAddr(),
		// Username: EnvRedisUsername(),    Username is for instances using Redis ACL. See more here: https://redis.io/docs/management/security/acl/
		Password: EnvRedisPassword(),
		DB:       db,
	})

	log.Println("Redis connection established...")
}
