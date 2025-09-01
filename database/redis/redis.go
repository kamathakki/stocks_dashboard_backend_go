package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"stock_automation_backend_go/shared/env"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func init() {
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%v:%v", env.GetEnv(env.EnvKeys.REDIS_HOST), env.GetEnv(env.EnvKeys.REDIS_PORT)),
		Password: env.GetEnv(env.EnvKeys.REDIS_PASSWORD),
		Username: env.GetEnv(env.EnvKeys.REDIS_USER),
		DB:       0,
	}

	redisClient = redis.NewClient(opts)

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
}

func GetKey(key string, result interface{}) (interface{}, error) {
	strResult, err := redisClient.Get(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	json.Unmarshal([]byte(strResult), &result)
	return result, nil
}

func SetKey(key string, value string) error {
	return redisClient.Set(context.Background(), key, value, 0).Err()
}

func DeleteKey(key string) error {
	return redisClient.Del(context.Background(), key).Err()
}

func GetHash(key string, field string, result interface{}) (interface{}, error) {
	strResult, err := redisClient.HGet(context.Background(), key, field).Result()
	if err != nil {
		return "", err
	}
	json.Unmarshal([]byte(strResult), &result)
	return redisClient.HGet(context.Background(), key, field).Result()
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
