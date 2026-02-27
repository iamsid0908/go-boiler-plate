package queue

import (
	"context"
	"log"

	"github.com/hibiken/asynq"
)

// StartWorker starts the asynq worker server.
// Pass a fully configured ServeMux with all task handlers registered.
// It starts processing in a background goroutine and returns the server for graceful shutdown.
func StartWorker(redisAddr string, mux *asynq.ServeMux) *asynq.Server {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			// Total concurrent workers across all queues
			Concurrency: 10,
			Queues: map[string]int{
				"webhooks":   6, // highest priority — webhook push events
				"default":    3, // medium — repo fetching
				"embeddings": 1, // low — embedding generation (rate-limited API)
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("[queue] ERROR processing task %s: %v", task.Type(), err)
			}),
		},
	)

	go func() {
		if err := srv.Run(mux); err != nil {
			log.Fatalf("[queue] Worker server failed: %v", err)
		}
	}()

	log.Println("[queue] Worker server started")
	return srv
}
