package broker

import (
	"context"
	"encoding/json"

	"github.com/jeromechua-12/task-queue/internal/models"
	"github.com/redis/go-redis/v9"
)

type Broker struct {
	redisClient *redis.Client
}

func New(client *redis.Client) *Broker {
	return &Broker{redisClient: client}
}

/*
Adds a Task to Redis priority queue for ready tasks.

First add mapping of task id to a Task serialised JSON object.
Then add the task id to redis priority queue.
*/
func (b *Broker) Enqueue(ctx context.Context, task models.Task) error {
	err := b.addTaskToSet(ctx, models.TaskHash, &task)
	if err != nil {
		return err
	}

	err = b.addIDToQueue(ctx, models.ReadyQueue, task.ID, task.Options.Priority)
	if err != nil {
		return err
	}

	return nil
}

// Adds a Task ID to Redis Sorted Set, scored by priority
func (b *Broker) addIDToQueue(ctx context.Context, queue string, id string, priority int) error {
	z := redis.Z{
		Score:  float64(priority),
		Member: id,
	}

	return b.redisClient.ZAdd(ctx, queue, z).Err()
}

// Adds mapping of Task ID to a Task in serialised JSON into a Redis Hash
func (b *Broker) addTaskToSet(ctx context.Context, hashKey string, task *models.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return b.redisClient.HSet(ctx, hashKey, task.ID, data).Err()
}
