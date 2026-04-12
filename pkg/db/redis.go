package db

import (
	"fmt"
	"time"

	"github.com/AleeCao/LogiTrack/pkg/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisConnection(env *config.EnvConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:           fmt.Sprintf("%s:%s", env.RedisHost, env.RedisPort),
		Password:       env.RedisPassword,
		DB:             0,
		Protocol:       2,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		MinIdleConns:   10,
		MaxIdleConns:   50,
		MaxActiveConns: 100,
	})

	return client
}
