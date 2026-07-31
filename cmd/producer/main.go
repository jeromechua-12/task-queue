package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
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
		err := broker.EnqueueTask(ctx, task)
		if err != nil {
			log.Fatal(err)
		}
	}

	// schedule sample tasks
	sampleScheduledTask := getScheduledSampleTasks()
	for _, task := range sampleScheduledTask {
		err := broker.ScheduleTask(ctx, task)
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
	var tasks []models.Task

	// generate 100 tasks with random priortity between 1-5
	for i := 0; i < 100; i++ {
		task := models.Task{
			ID:        fmt.Sprintf("uuid-%d", i),
			Name:      "some_function",
			CreatedAt: time.Now().UTC(),
			Options: models.TaskOptions{
				Priority:      rand.IntN(5) + 1,
				RetryDelay:    0,
				MaxRetries:    3,
				TotalAttempts: 0,
			},
		}
		tasks = append(tasks, task)
	}
	return tasks
}

func getScheduledSampleTasks() []models.Task {
	var tasks []models.Task

	// generate 50 tasks with random priority between 1-5 and random delays 
	for i := 0; i < 50; i++ {
		task := models.Task{
			ID:        fmt.Sprintf("uuid-scheduled-%d", i),
			Name:      "some_function",
			CreatedAt: time.Now().UTC(),
			Options: models.TaskOptions{
				Priority:      rand.IntN(5) + 1,
				RetryDelay:    0,
				MaxRetries:    3,
				TotalAttempts: 0,
			},
		}
		// generate random delay in seconds betwen 30 mins and 2 days
		randMinDelay := int64(time.Duration(30) * time.Minute / time.Second) // 30 mins
		randMaxDelay := int64(time.Duration(48) * time.Hour / time.Second)   // 2 days
		randDelay := rand.Int64N(randMaxDelay-randMinDelay) + randMinDelay
		task.Options.Delay = randDelay

		tasks = append(tasks, task)
	}

	return tasks
}
