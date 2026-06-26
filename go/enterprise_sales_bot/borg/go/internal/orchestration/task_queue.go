package orchestration

/**
 * @file task_queue.go
 * @module go/internal/orchestration
 *
 * WHAT: Go-native implementation of the collective task queue.
 * Allows agents to publish and claim background tasks.
 */

import (
	"fmt"
	"sort"
	"sync"
)

type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskCompleted  TaskStatus = "completed"
	TaskFailed     TaskStatus = "failed"
)

type A2ATask struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Priority    int        `json:"priority"`
	Status      TaskStatus `json:"status"`
	AssignedTo  string     `json:"assignedTo,omitempty"`
	Result      any        `json:"result,omitempty"`
	Error       string     `json:"error,omitempty"`
}

type TaskQueue struct {
	mu     sync.RWMutex
	tasks  map[string]*A2ATask
	broker *A2ABroker
}

func NewTaskQueue(broker *A2ABroker) *TaskQueue {
	return &TaskQueue{
		tasks:  make(map[string]*A2ATask),
		broker: broker,
	}
}

func (q *TaskQueue) PublishTask(task A2ATask) {
	q.mu.Lock()
	task.Status = TaskPending
	q.tasks[task.ID] = &task
	q.mu.Unlock()

	fmt.Printf("[Go TaskQueue] 📥 New task queued: %s (P%d)\n", task.Description, task.Priority)

	// Broadcast notice
	q.broker.RouteMessage(A2AMessage{
		ID:        fmt.Sprintf("work-%s", task.ID),
		Timestamp: nowMillis(),
		Sender:    "TASK_QUEUE",
		Type:      TaskRequest,
		Payload: map[string]interface{}{
			"task":     task.Description,
			"taskId":   task.ID,
			"priority": task.Priority,
		},
	})
}

func (q *TaskQueue) ClaimTask(taskID, agentID string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, ok := q.tasks[taskID]
	if ok && task.Status == TaskPending {
		task.Status = TaskInProgress
		task.AssignedTo = agentID
		fmt.Printf("[Go TaskQueue] 🤝 Task %s claimed by %s\n", taskID, agentID)
		return true
	}
	return false
}

func (q *TaskQueue) ListTasks() []A2ATask {
	q.mu.RLock()
	defer q.mu.RUnlock()

	list := make([]A2ATask, 0, len(q.tasks))
	for _, t := range q.tasks {
		list = append(list, *t)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Priority > list[j].Priority
	})

	return list
}
