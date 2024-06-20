package worker

import (
	"context"
	"log"
)

func RunWorker(ctx context.Context, nWorkers int, done chan struct{}) {
	log.Printf("Starting %d workers\n", nWorkers)
}
