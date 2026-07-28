package broker

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/jeromechua-12/task-queue/internal/models"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6380",
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(context.TODO()).Result()
	if err != nil {
		t.Fatal(err)
	}

	return rdb
}

func cleanTestRedis(t *testing.T, rdb *redis.Client) {
	t.Helper()
	rdb.FlushAll(context.TODO())
}

func TestEnqueueDequeue(t *testing.T) {
	client := newTestRedis(t)
	t.Cleanup(func() {
		cleanTestRedis(t, client)
	})

	ctx := context.TODO()
	broker := Broker{redisClient: client}

	tasks := []models.Task{
		{
			ID:        "1",
			Name:      "send_email",
			CreatedAt: time.Now().UTC(),
			Options: models.TaskOptions{
				Priority:      2,
				Delay:         0,
				MaxRetries:    3,
				TotalAttempts: 0,
			},
		},
		{
			ID:        "2",
			Name:      "send_email",
			CreatedAt: time.Now().UTC(),
			Options: models.TaskOptions{
				Priority:      1,
				Delay:         0,
				MaxRetries:    3,
				TotalAttempts: 0,
			},
		},
		{
			ID:        "3",
			Name:      "send_email",
			CreatedAt: time.Now().UTC(),
			Options: models.TaskOptions{
				Priority:      2,
				Delay:         0,
				MaxRetries:    3,
				TotalAttempts: 0,
			},
		},
	}

	for _, task := range tasks {
		err := broker.Enqueue(ctx, task)
		if err != nil {
			t.Fatal(err)
		}
	}

	expectedDequeueOrder := [3]int{1, 0, 2}
	for _, idx := range expectedDequeueOrder {
		task, err := broker.DequeueTask(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if task == nil {
			t.Errorf("got nil; want %v", tasks[idx])
		}

		if equal := reflect.DeepEqual(*task, tasks[idx]); !equal {
			t.Errorf("got %v; want %v", *task, tasks[idx])
		}
	}

	// next dequeue should return nil
	task, err := broker.DequeueTask(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if task != nil {
		t.Errorf("got %v; want nil", *task)
	}
}
