package main

import (
	"context"
	"log"
	"os"

	"github.com/jeromechua-12/task-queue/internal/broker"
	"github.com/jeromechua-12/task-queue/internal/worker"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	ctx := context.Background()

	redisAddress := os.Getenv("REDIS_ADDRESS")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisOptions := redis.Options{
		Addr:     redisAddress,
		Password: redisPassword,
		DB:       0, // use default DB
	}

	redisClient, err := openRedis(ctx, &redisOptions)
	if err != nil {
		log.Fatalf("error loading redis: %v\n", err)
	}

	// initialise broker
	broker := broker.New(redisClient)

	// initialise worker
	worker := worker.New(broker)
	worker.WaitOrExecuteTask(ctx)
}

func openRedis(ctx context.Context, options *redis.Options) (*redis.Client, error) {
	rdb := redis.NewClient(options)
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}
