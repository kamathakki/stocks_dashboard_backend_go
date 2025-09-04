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

func init() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := &redis.Options{
		Addr:     fmt.Sprintf("%v:%v", env.GetEnv(env.EnvKeys.REDIS_HOST), env.GetEnv(env.EnvKeys.REDIS_PORT)),
		Password: env.GetEnv(env.EnvKeys.REDIS_PASSWORD),
		Username: env.GetEnv(env.EnvKeys.REDIS_USER),
		DB:       0,
	}

	redisClient = redis.NewClient(opts)

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("redis unavailable %v; continuing without cache", err)
		redisClient = nil
	}
}

func GetKey[T any](key string) (*T, error) {
	result := new(T)
	strResult, err := redisClient.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(strResult), result); err != nil {
		return nil, err
	}
	return result, nil
}

func SetKey(key string, value string) error {
	return redisClient.Set(context.Background(), key, value, 0).Err()
}

func DeleteKey(key string) error {
	return redisClient.Del(context.Background(), key).Err()
}

func GetHash[T any](key string, field string) (*T, error) {
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
	return redisClient.HSet(context.Background(), key, field, value).Err()
}

func DeleteHash(key string, field string) error {
	return redisClient.HDel(context.Background(), key, field).Err()
}

func QuitRedis() error {
	return redisClient.Close()
}
