package worker

import (
	"context"
	"errors"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/jeromechua-12/task-queue/internal/broker"
	"github.com/jeromechua-12/task-queue/internal/models"
)

type Worker struct {
	broker *broker.Broker
}

func New(broker *broker.Broker) *Worker {
	return &Worker{broker: broker}
}

/*
Waits for a task in the queue. If there is a task, pops and
executes it.

If there is an error during execution and total attempts less than max attempts,
put it into queue to retry with a delay. Else, put it in deadletter queue.
*/
func (w *Worker) WaitOrExecuteTask(ctx context.Context) {
	for {
		task, err := w.broker.DequeueTask(ctx)

		// Log error about redis operation
		if err != nil {
			log.Printf("Error: %v\n", err)
		}

		if task == nil {
			continue
		}

		err = w.executeTask(task)

		// handle task execution error
		if err != nil {
			if task.Options.TotalAttempts < task.Options.MaxRetries {
				// increment task total attempts and update hash
				task.Options.TotalAttempts++
				w.broker.AddTaskToHash(ctx, models.TaskHash, task)
				// add to scheduled queue
				w.retryTask(ctx, task)
			} else {
				// TODO: Add to DLQ
			}
			continue
		}

		// #TODO: if task succeeds, remove from set
	}
}

// Schedules a failed task for retry.
func (w *Worker) retryTask(ctx context.Context, task *models.Task) {
	scheduledTime := getDelayTime(task)
	w.broker.RetryTask(ctx, task, scheduledTime)
}

/*
Returns execution time for delayed task in UTC Unix timestamp.

Apply exponential delay with jitters based on number of retries attempted.
*/
func getDelayTime(task *models.Task) int64 {
	curTime := time.Now().UTC().Unix()

	curAttempts := task.Options.TotalAttempts
	delay := task.Options.Delay

	// random jitter of 0-5 seconds
	jitter := rand.Int63n(5)
	totalDelay := math.Pow(float64(delay), float64(curAttempts)) + float64(jitter)

	scheduledTime := curTime + int64(totalDelay)

	return scheduledTime
}


/*
Executes a tasks with random chance of failure.
*/
func (w *Worker) executeTask(task *models.Task) error {
	log.Printf("Executing task: (ID:%s) %s", task.ID, task.Name)

	// random execution time between 1-2 seconds
	seconds := rand.Float32() + 1
	time.Sleep(time.Duration(seconds) * time.Second)
	
	// random number between 0-4. If 0, return error
	n := rand.Intn(5)

	if n == 0 {
		log.Printf("Task error: (ID:%s) %s", task.ID, task.Name)
		return errors.New("Task execution error")
	}

	return nil
}
