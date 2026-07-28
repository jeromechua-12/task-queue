package broker

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jeromechua-12/task-queue/internal/models"
	"github.com/redis/go-redis/v9"
)

// embed lua scripts
//
//go:embed scripts
var fs embed.FS

// scripts path
const (
	DequeueTaskScript string = "scripts/get_task.lua"
)

type Broker struct {
	redisClient *redis.Client
}

func New(client *redis.Client) *Broker {
	return &Broker{redisClient: client}
}

/*
Adds a Task to Redis Sorted Set for tasks that are ready to be processed.

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

/*
Dequeues task from Redis Sorted Set storing tasks that are ready to be processed.

Perform Redis Eval on lua script that calls redis ZPOPMIN and HGET.
Lua sceripts guarentees atomic execution.

Returns a pointer to a Task.
*/
func (b *Broker) DequeueTask(ctx context.Context) (*models.Task, error) {
	// read lua script
	script, err := readScript(DequeueTaskScript)
	if err != nil {
		return nil, err
	}

	// pass in Redis keys in the required order: Sorted Set, Hash
	keys := []string{models.ReadyQueue, models.TaskHash}

	result, err := b.redisClient.Eval(ctx, script, keys).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var task models.Task

	data, ok := result.(string)
	if !ok {
		return nil, errors.New("Dequeue error: Type assertion of Task failed")
	}

	err = json.Unmarshal([]byte(data), &task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// Adds a Task id to Redis Sorted Set, scored by priority
func (b *Broker) addIDToQueue(ctx context.Context, queue string, id string, priority int) error {
	z := redis.Z{
		Score:  float64(priority),
		Member: id,
	}

	return b.redisClient.ZAdd(ctx, queue, z).Err()
}

// Maps Task ID to a Task in serialised JSON in a Redis Hash
func (b *Broker) addTaskToSet(ctx context.Context, hashKey string, task *models.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return b.redisClient.HSet(ctx, hashKey, task.ID, data).Err()
}

// Reads lua scripts and return file content
func readScript(filepath string) (string, error) {
	bytes, err := fs.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytes)), nil
}
