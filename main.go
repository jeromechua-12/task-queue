package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	ctx := context.Background()

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisOptions := redis.Options{
		Addr:     "localhost:6379",
		Password: redisPassword,
		DB:       0,
	}

	redisClient, err := openRedis(ctx, &redisOptions)
	if err != nil {
		log.Fatalf("error loading redis: %v\n", err)
	}

	log.Println("redis client started")
	pong, _ := redisClient.Ping(ctx).Result()
	log.Println(pong)
}

func openRedis(ctx context.Context, options *redis.Options) (*redis.Client, error) {
	rdb := redis.NewClient(options)
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}
