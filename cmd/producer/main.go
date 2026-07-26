package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jeromechua-12/task-queue/internal/broker"
	"github.com/jeromechua-12/task-queue/internal/models"

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

	// enqueue sample tasks
	sampleTasks := getSampleTasks()
	for _, task := range sampleTasks {
		err := broker.Enqueue(ctx, task)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func openRedis(ctx context.Context, options *redis.Options) (*redis.Client, error) {
	rdb := redis.NewClient(options)
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}

func getSampleTasks() []models.Task {
	tasks := []models.Task{
		{
			ID:        "uuid-1",
			Name:      "function_1",
			Status:    models.StatusPending,
			CreatedAt: time.Now().UTC(),
			Options: models.TaskOptions{
				Priority:      2,
				Delay:         0,
				MaxRetries:    3,
				TotalAttempts: 0,
			},
		},
		{
			ID:        "uuid-2",
			Name:      "function_2",
			Status:    models.StatusPending,
			CreatedAt: time.Now().UTC(),
			Options: models.TaskOptions{
				Priority:      1,
				Delay:         0,
				MaxRetries:    3,
				TotalAttempts: 0,
			},
		},
		{
			ID:        "uuid-3",
			Name:      "function_3",
			Status:    models.StatusPending,
			CreatedAt: time.Now().UTC(),
			Options: models.TaskOptions{
				Priority:      2,
				Delay:         int(30 * time.Minute / time.Second),
				MaxRetries:    3,
				TotalAttempts: 0,
			},
		},
		{
			ID:        "uuid-4",
			Name:      "function_4",
			Status:    models.StatusPending,
			CreatedAt: time.Now().UTC(),
			Options: models.TaskOptions{
				Priority:      3,
				Delay:         int(3 * time.Hour / time.Second),
				MaxRetries:    3,
				TotalAttempts: 0,
			},
		},
	}
	return tasks
}
