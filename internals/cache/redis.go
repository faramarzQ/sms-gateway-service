package cache

import (
	"context"
	"github.com/faramarzQ/sms-gateway-service/internals/config"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	return client
}
