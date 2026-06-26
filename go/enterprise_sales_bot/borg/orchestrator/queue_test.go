package orchestrator

import (
	"os"
	"testing"
)

func TestQueueInitialization(t *testing.T) {
	dbPath := "./.test_tormentnexus_queue.db"

	// Ensure clean state
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	q, err := NewTaskQueue(dbPath)
	if err != nil {
		t.Fatalf("Failed to spin up local SQLite BullMQ parity queue: %v", err)
	}

	err = q.Enqueue("TEST_PAYLOAD_ECHO")
	if err != nil {
		t.Errorf("Failed to enqueue payload onto native SQLite DB: %v", err)
	}

	// Validate status
	var id int
	var payload, status string
	err = q.db.QueryRow("SELECT id, payload, status FROM job_queue LIMIT 1").Scan(&id, &payload, &status)
	if err != nil {
		t.Errorf("Queue schema validation mismatch: %v", err)
	}

	if payload != "TEST_PAYLOAD_ECHO" || status != "pending" {
		t.Errorf("SQLite database schema hallucinated fields. Expected TEST_PAYLOAD_ECHO/pending, got %s/%s", payload, status)
	}

	q.Close()
}
