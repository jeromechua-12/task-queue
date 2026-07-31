package worker

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/jeromechua-12/task-queue/internal/broker"
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

		log.Printf("Executing task: %s", task.ID)

		// Random execution time between 1-2 seconds
		seconds := rand.Float32() + 1
		time.Sleep(time.Duration(seconds) * time.Second)
	}
}
