package store_test

import (
	"testing"
	"time"

	"github.com/comma-compliance/arc-relay/internal/store"
	"github.com/comma-compliance/arc-relay/internal/testutil"
)

func newQueueItem(serverID string) *store.ArchiveQueueItem {
	return &store.ArchiveQueueItem{
		ServerID:     serverID,
		Payload:      `{"version":"v1","source":"arc_relay","phase":"test"}`,
		URL:          "https://compliance.example.com/ingest",
		AuthType:     "bearer",
		AuthValue:    "test-token",
		APIKeyHeader: "X-API-Key",
	}
}

func TestEnqueue_InsertsRow(t *testing.T) {
	db := testutil.OpenTestDB(t)
	qs := store.NewArchiveQueueStore(db, nil)

	item := newQueueItem("srv-test")
	if err := qs.Enqueue(item); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if item.ID == "" {
		t.Error("expected ID to be set after enqueue")
	}

	items, err := qs.DequeueDue(10)
	if err != nil {
		t.Fatalf("DequeueDue: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Payload != item.Payload {
		t.Errorf("payload = %s, want %s", items[0].Payload, item.Payload)
	}
	if items[0].URL != "https://compliance.example.com/ingest" {
		t.Errorf("url = %s", items[0].URL)
	}
	if items[0].AuthType != "bearer" {
		t.Errorf("auth_type = %s", items[0].AuthType)
	}
	if items[0].Status != "pending" {
		t.Errorf("status = %s, want pending", items[0].Status)
	}
}

func TestDequeueDue_RespectsNextAttemptAt(t *testing.T) {
	db := testutil.OpenTestDB(t)
	qs := store.NewArchiveQueueStore(db, nil)

	// Enqueue one due now
	item1 := newQueueItem("srv-1")
	if err := qs.Enqueue(item1); err != nil {
		t.Fatal(err)
	}

	// Enqueue another and reschedule it into the future
	item2 := newQueueItem("srv-2")
	if err := qs.Enqueue(item2); err != nil {
		t.Fatal(err)
	}
	if err := qs.Reschedule(item2.ID, time.Now().Add(1*time.Hour), "test future"); err != nil {
		t.Fatal(err)
	}

	items, err := qs.DequeueDue(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 due item, got %d", len(items))
	}
	if items[0].ID != item1.ID {
		t.Errorf("expected item1, got %s", items[0].ID)
	}
}

func TestDequeueDue_RespectsLimit(t *testing.T) {
	db := testutil.OpenTestDB(t)
	qs := store.NewArchiveQueueStore(db, nil)

	for i := 0; i < 5; i++ {
		if err := qs.Enqueue(newQueueItem("srv-1")); err != nil {
			t.Fatal(err)
		}
	}

	items, err := qs.DequeueDue(3)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
}

func TestDequeueDue_OrdersByNextAttemptThenCreated(t *testing.T) {
	db := testutil.OpenTestDB(t)
	qs := store.NewArchiveQueueStore(db, nil)

	item1 := newQueueItem("srv-1")
	if err := qs.Enqueue(item1); err != nil {
		t.Fatal(err)
	}
	item2 := newQueueItem("srv-2")
	if err := qs.Enqueue(item2); err != nil {
		t.Fatal(err)
	}

	items, err := qs.DequeueDue(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) < 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != item1.ID {
		t.Error("expected item1 first (earlier created_at)")
	}
}

func TestMarkDelivered_DeletesRow(t *testing.T) {
	db := testutil.OpenTestDB(t)
	qs := store.NewArchiveQueueStore(db, nil)

	item := newQueueItem("srv-1")
	if err := qs.Enqueue(item); err != nil {
		t.Fatal(err)
	}

	if err := qs.MarkDelivered(item.ID); err != nil {
		t.Fatalf("MarkDelivered: %v", err)
	}

	items, _ := qs.DequeueDue(10)
	if len(items) != 0 {
		t.Errorf("expected 0 items after delivery, got %d", len(items))
	}
}

func TestReschedule_UpdatesAttemptsAndNextAttempt(t *testing.T) {
	db := testutil.OpenTestDB(t)
	qs := store.NewArchiveQueueStore(db, nil)

	item := newQueueItem("srv-1")
	if err := qs.Enqueue(item); err != nil {
		t.Fatal(err)
	}

	future := time.Now().Add(5 * time.Minute)
	if err := qs.Reschedule(item.ID, future, "connection timeout"); err != nil {
		t.Fatalf("Reschedule: %v", err)
	}

	// Should not be due yet
	items, _ := qs.DequeueDue(10)
	if len(items) != 0 {
		t.Error("rescheduled item should not be due yet")
	}

	// Check status shows it
	st, err := qs.Status()
	if err != nil {
		t.Fatal(err)
	}
	if st.PendingCount != 1 {
		t.Errorf("pending = %d, want 1", st.PendingCount)
	}
	if st.LastError != "connection timeout" {
		t.Errorf("last_error = %s", st.LastError)
	}
}

func TestMarkHold_SetsStatusHold(t *testing.T) {
	db := testutil.OpenTestDB(t)
	qs := store.NewArchiveQueueStore(db, nil)

	item := newQueueItem("srv-1")
	if err := qs.Enqueue(item); err != nil {
		t.Fatal(err)
	}

	if err := qs.MarkHold(item.ID, "400 Bad Request"); err != nil {
		t.Fatalf("MarkHold: %v", err)
	}

	// Should not appear in due items
	items, _ := qs.DequeueDue(10)
	if len(items) != 0 {
		t.Error("held item should not be dequeued")
	}

	st, _ := qs.Status()
	if st.HoldCount != 1 {
		t.Errorf("hold = %d, want 1", st.HoldCount)
	}
	if st.PendingCount != 0 {
		t.Errorf("pending = %d, want 0", st.PendingCount)
	}
}

func TestRetryHeld_ResetsHeldRows(t *testing.T) {
	db := testutil.OpenTestDB(t)
	qs := store.NewArchiveQueueStore(db, nil)

	item := newQueueItem("srv-1")
	if err := qs.Enqueue(item); err != nil {
		t.Fatal(err)
	}
	if err := qs.MarkHold(item.ID, "permanent error"); err != nil {
		t.Fatal(err)
	}

	count, err := qs.RetryHeld()
	if err != nil {
		t.Fatalf("RetryHeld: %v", err)
	}
	if count != 1 {
		t.Errorf("retried = %d, want 1", count)
	}

	items, _ := qs.DequeueDue(10)
	if len(items) != 1 {
		t.Fatalf("expected 1 item after retry, got %d", len(items))
	}
	if items[0].Attempts != 0 {
		t.Errorf("attempts should be reset to 0, got %d", items[0].Attempts)
	}
}

func TestStatus_ReturnsCorrectCounts(t *testing.T) {
	db := testutil.OpenTestDB(t)
	qs := store.NewArchiveQueueStore(db, nil)

	// Empty queue
	st, err := qs.Status()
	if err != nil {
		t.Fatal(err)
	}
	if st.TotalCount != 0 {
		t.Errorf("empty queue total = %d", st.TotalCount)
	}

	// Add items in various states
	item1 := newQueueItem("srv-1")
	if err := qs.Enqueue(item1); err != nil {
		t.Fatal(err)
	}

	item2 := newQueueItem("srv-1")
	if err := qs.Enqueue(item2); err != nil {
		t.Fatal(err)
	}
	if err := qs.MarkHold(item2.ID, "bad request"); err != nil {
		t.Fatal(err)
	}

	item3 := newQueueItem("srv-2")
	if err := qs.Enqueue(item3); err != nil {
		t.Fatal(err)
	}
	if err := qs.Reschedule(item3.ID, time.Now().Add(1*time.Hour), "timeout"); err != nil {
		t.Fatal(err)
	}

	st, _ = qs.Status()
	if st.PendingCount != 2 { // item1 (due) + item3 (pending but future)
		t.Errorf("pending = %d, want 2", st.PendingCount)
	}
	if st.HoldCount != 1 {
		t.Errorf("hold = %d, want 1", st.HoldCount)
	}
	if st.TotalCount != 3 {
		t.Errorf("total = %d, want 3", st.TotalCount)
	}
}

func TestStatusForServer_FiltersCorrectly(t *testing.T) {
	db := testutil.OpenTestDB(t)
	qs := store.NewArchiveQueueStore(db, nil)

	if err := qs.Enqueue(newQueueItem("srv-1")); err != nil {
		t.Fatal(err)
	}
	if err := qs.Enqueue(newQueueItem("srv-1")); err != nil {
		t.Fatal(err)
	}
	if err := qs.Enqueue(newQueueItem("srv-2")); err != nil {
		t.Fatal(err)
	}

	st1, _ := qs.StatusForServer("srv-1")
	if st1.PendingCount != 2 {
		t.Errorf("srv-1 pending = %d, want 2", st1.PendingCount)
	}

	st2, _ := qs.StatusForServer("srv-2")
	if st2.PendingCount != 1 {
		t.Errorf("srv-2 pending = %d, want 1", st2.PendingCount)
	}
}

func TestPrune_RemovesOldItems(t *testing.T) {
	db := testutil.OpenTestDB(t)
	qs := store.NewArchiveQueueStore(db, nil)

	// Insert an item with old timestamp
	item := newQueueItem("srv-1")
	if err := qs.Enqueue(item); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE archive_queue SET created_at = datetime('now', '-8 days') WHERE id = ?`, item.ID); err != nil {
		t.Fatal(err)
	}

	// Insert a recent item
	if err := qs.Enqueue(newQueueItem("srv-1")); err != nil {
		t.Fatal(err)
	}

	pruned, err := qs.Prune(7 * 24 * time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if pruned != 1 {
		t.Errorf("pruned = %d, want 1", pruned)
	}

	st, _ := qs.Status()
	if st.TotalCount != 1 {
		t.Errorf("remaining = %d, want 1", st.TotalCount)
	}
}

func TestDenormalization_PreservesOriginalConfig(t *testing.T) {
	db := testutil.OpenTestDB(t)
	qs := store.NewArchiveQueueStore(db, nil)

	item := &store.ArchiveQueueItem{
		ServerID:     "srv-1",
		Payload:      `{"test":true}`,
		URL:          "https://original.example.com/ingest",
		AuthType:     "api_key",
		AuthValue:    "original-key",
		APIKeyHeader: "X-Custom-Key",
	}
	if err := qs.Enqueue(item); err != nil {
		t.Fatal(err)
	}

	items, _ := qs.DequeueDue(1)
	if len(items) != 1 {
		t.Fatal("expected 1 item")
	}
	if items[0].URL != "https://original.example.com/ingest" {
		t.Errorf("URL changed: %s", items[0].URL)
	}
	if items[0].AuthType != "api_key" {
		t.Errorf("AuthType changed: %s", items[0].AuthType)
	}
	if items[0].APIKeyHeader != "X-Custom-Key" {
		t.Errorf("APIKeyHeader changed: %s", items[0].APIKeyHeader)
	}
}
