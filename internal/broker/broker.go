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
Adds a Task that is ready to be processed into Redis Sorted Set.
Tasks are sorted by priority.

First add mapping of task id to a serialised Task into a Redis Hash.
Then add the task id to Redis Sorted Set.
*/
func (b *Broker) EnqueueTask(ctx context.Context, task models.Task) error {
	err := b.addTaskToHash(ctx, models.TaskHash, &task)
	if err != nil {
		return err
	}

	err = b.addIDToQueue(ctx, models.ReadyQueue, task.ID, float64(task.Options.Priority))
	if err != nil {
		return err
	}

	return nil
}

/*
Pops a Task that is ready to be processed from Redis Sorted Set.
Tasks are popped by priority.

Perform Redis Eval on lua script that calls redis ZPOPMIN and HGET.
Lua scripts guarentees atomic execution.

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

/*
Adds a Task scheduled for a later execution into Redis Sorted Set.
Scheduled time is calculated by created time + delay.

Tasks are sorted by execution time in UTC Unix timestamp.

First add mapping of task id to a serialised Task into a Redis Hash.
Then add the task id to Redis Sorted Set.
*/
func (b *Broker) ScheduleTask(ctx context.Context, task models.Task) error {
	executionTime := task.CreatedAt.UTC().Unix() + task.Options.Delay

	err := b.addTaskToHash(ctx, models.TaskHash, &task)
	if err != nil {
		return err
	}

	err = b.addIDToQueue(ctx, models.DelayQueue, task.ID, float64(executionTime))
	if err != nil {
		return err
	}

	return nil
}

// Adds a Task for retry. Task is added into the same Redis Sorted Set as other scheduled tasks.
func (b *Broker) RetryTask(ctx context.Context, task models.Task, retryTime int64) error {
	err := b.addIDToQueue(ctx, models.DelayQueue, task.ID, float64(retryTime))
	if err != nil {
		return err
	}

	return nil
}

// Adds a Task id to Redis Sorted Set, scored by priority.
func (b *Broker) addIDToQueue(ctx context.Context, queue string, id string, priority float64) error {
	z := redis.Z{
		Score:  priority,
		Member: id,
	}

	return b.redisClient.ZAdd(ctx, queue, z).Err()
}

// Adds a mapping of Task ID to a serialised Task into a Redis Hash.
func (b *Broker) addTaskToHash(ctx context.Context, hashKey string, task *models.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return b.redisClient.HSet(ctx, hashKey, task.ID, data).Err()
}

// Reads lua scripts and return file content.
func readScript(filepath string) (string, error) {
	bytes, err := fs.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytes)), nil
}
