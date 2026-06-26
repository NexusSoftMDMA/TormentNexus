package agent

import (
	"fmt"
	"time"
)

// AsyncWorker handles background task execution (Jules parity)
type AsyncWorker struct {
	tasks chan string
}

func NewAsyncWorker() *AsyncWorker {
	worker := &AsyncWorker{
		tasks: make(chan string, 100),
	}
	go worker.start()
	return worker
}

func (w *AsyncWorker) Enqueue(task string) {
	w.tasks <- task
	fmt.Printf("[Async] Task enqueued: %s\n", task)
}

func (w *AsyncWorker) start() {
	for task := range w.tasks {
		// Simulate long-running background execution
		time.Sleep(2 * time.Second)
		fmt.Printf("\n[Async] Background task completed: %s\n", task)
	}
}
