package config

import (
	"github.com/go-redis/redis/v8"
)

var Redis *redis.Client

func CreateRedisClient() {
	// opt, err := redis.ParseURL("redis://localhost:6364/0")
	// if err != nil {
	// 	log.Print("CreateRedisClient", err)
	// 	panic(err)
	// }
	// redis := redis.NewClient(opt)
	// Redis = redis

	redis := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	Redis = redis
}
