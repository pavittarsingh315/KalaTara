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

const (
	cacheQueryTimeout = 500 * time.Millisecond
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

	ctx, cancel := NewCacheContext()
	defer cancel()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Error connecting to cache...")
		panic(err)
	}

	log.Println("Redis connection established...")
}

// Returns a new context with a timeout of 500 milliseconds
func NewCacheContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), cacheQueryTimeout)
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

// Delete key(s) from cache. Returns true if all key(s) deleted successfully, else false
func Delete(ctx context.Context, keys ...string) bool {
	num_keys_removed, err := rdb.Del(ctx, keys...).Result()
	if err != nil {
		return false
	}
	return num_keys_removed == int64(len(keys))
}

// Returns the remaining duration of the key's lifespan
func ExpiresIn(ctx context.Context, key string) (time.Duration, error) {
	duration, err := rdb.TTL(ctx, key).Result()
	if err != nil {
		return time.Second * 0, err
	}
	return duration, nil
}
