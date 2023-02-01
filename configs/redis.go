package configs

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	rdb *redis.Client
)

func InitRedis() {
	db, err := strconv.Atoi(EnvRedisDatabase())
	if err != nil {
		log.Fatal("Error connecting to Redis...")
		panic(err)
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: EnvRedisAddr(),
		// Username: EnvRedisUsername(),    Username is for instances using Redis ACL. See more here: https://redis.io/docs/management/security/acl/
		Password: EnvRedisPassword(),
		DB:       db,
	})

	log.Println("Redis connection established...")
}

// Sets a new key. Zero expiration means the key has no expiration time
func RedisSet(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := rdb.Set(ctx, key, val, expiration).Err(); err != nil {
		return err
	}

	return nil
}

// Gets value for a key if it exists.
//
// dest must be a pointer.
func RedisGet(ctx context.Context, key string, dest interface{}) error {
	value, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), dest)
}

// Returns the formatted key for profile objects given the user id.
func RedisProfileKey(user_id string) string {
	return "user:" + user_id + ":profile"
}
