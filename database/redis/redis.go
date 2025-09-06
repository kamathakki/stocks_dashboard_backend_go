package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"stock_automation_backend_go/shared/env"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func InitRedis() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := &redis.Options{
		Addr:     fmt.Sprintf("%v:%v", env.GetEnv[string](env.EnvKeys.REDIS_HOST), env.GetEnv[string](env.EnvKeys.REDIS_PORT)),
		Password: env.GetEnv[string](env.EnvKeys.REDIS_PASSWORD),
		Username: env.GetEnv[string](env.EnvKeys.REDIS_USER),
		DB:       0,
	}

	redisClient = redis.NewClient(opts)

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("redis unavailable %v; continuing without cache", err)
		redisClient = nil
	}
}

func GetKey[T any](key string) (*T, error) {
	// If redis is unavailable, behave like a cache miss
	if redisClient == nil {
		return new(T), redis.Nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := new(T)
	strResult, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal([]byte(strResult), result); err != nil {
		return result, err
	}
	return result, nil
}

func SetKey(key string, value string) error {
	if redisClient == nil {
		return nil
	}
	return redisClient.Set(context.Background(), key, value, 0).Err()
}

func DeleteKey(key string) error {
	if redisClient == nil {
		return nil
	}
	return redisClient.Del(context.Background(), key).Err()
}

func GetHash[T any](key string, field string) (*T, error) {
	if redisClient == nil {
		fmt.Println("Redis is unavailable")
		return new(T), redis.Nil
	}
	result := new(T)
	strResult, err := redisClient.HGet(context.Background(), key, field).Result()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(strResult), result); err != nil {
		return nil, err
	}
	return result, nil
}

func SetHash(key string, field string, value string) error {
	if redisClient == nil {
		return nil
	}
	return redisClient.HSet(context.Background(), key, field, value).Err()
}

func DeleteHash(key string, field string) error {
	if redisClient == nil {
		return nil
	}
	return redisClient.HDel(context.Background(), key, field).Err()
}

func QuitRedis() error {
	if redisClient == nil {
		return nil
	}
	return redisClient.Close()
}
