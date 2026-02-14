package core

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/blcvn/backend/services/pkg/queue"
	"github.com/hibiken/asynq"
)

type TaskDispatcher struct {
	producer *queue.Producer
}

func NewTaskDispatcher() *TaskDispatcher {
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	cfg := queue.RedisConfig{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	}
	return &TaskDispatcher{
		producer: queue.NewProducer(cfg),
	}
}

func (d *TaskDispatcher) DispatchTask(ctx context.Context, taskType string, payload interface{}, priority string) error {
	var opts []asynq.Option

	switch priority {
	case "critical":
		opts = append(opts, asynq.Queue("critical"))
	case "low":
		opts = append(opts, asynq.Queue("low"))
	default:
		opts = append(opts, asynq.Queue("default"))
	}

	info, err := d.producer.Enqueue(ctx, taskType, payload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %s: %w", taskType, err)
	}

	fmt.Printf("Enqueued task: id=%s queue=%s\n", info.ID, info.Queue)
	return nil
}

func (d *TaskDispatcher) Close() {
	d.producer.Close()
}
