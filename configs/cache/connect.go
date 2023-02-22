package cache

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"nerajima.com/NeraJima/configs"
)

var (
	rdb *redis.Client
)

func Initialize() {
	db, err := strconv.Atoi(configs.EnvRedisDatabase())
	if err != nil {
		log.Fatal("Error connecting to Redis...")
		panic(err)
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: configs.EnvRedisAddr(),
		// Username: configs.EnvRedisUsername(),    Username is for instances using Redis ACL. See more here: https://redis.io/docs/management/security/acl/
		Password: configs.EnvRedisPassword(),
		DB:       db,
	})

	log.Println("Redis connection established...")
}

// Gets value for a key if it exists. dest must be a pointer.
func Get(ctx context.Context, key string, dest interface{}) error {
	value, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), dest)
}

// Sets a new key. Zero expiration means the key has no expiration time
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if err := rdb.Set(ctx, key, val, expiration).Err(); err != nil {
		return err
	}
	return nil
}
