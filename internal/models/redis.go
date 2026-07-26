package models

const (
	ReadyQueue      string = "taskqueue:ready"
	ProcessingQueue string = "taskqueue:processing"
	DelayQueue      string = "taskqueue:delay"
	DeadLetterQueue string = "taskqueue:deadletter"
	TaskHash        string = "taskhash"
)
