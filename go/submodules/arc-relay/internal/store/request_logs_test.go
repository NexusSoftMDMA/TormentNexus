package store_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/comma-compliance/arc-relay/internal/store"
	"github.com/comma-compliance/arc-relay/internal/testutil"
)

func insertTestServerAndUser(t *testing.T, db *store.DB, serverID, serverName, userID, username string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO servers (id, name, display_name, server_type, config, status, created_at, updated_at)
		VALUES (?, ?, ?, 'stdio', '{}', 'stopped', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		serverID, serverName, serverName)
	if err != nil {
		t.Fatalf("inserting server %s: %v", serverID, err)
	}

	users := store.NewUserStore(db)
	_, err = users.CreateWithAccessLevel(username, "pass", "user", "write", nil)
	if err != nil {
		t.Fatalf("creating user %s: %v", username, err)
	}
	// Fetch the actual ID to update the caller's expected userID
	u, err := users.GetByUsername(username)
	if err != nil || u == nil {
		t.Fatalf("fetching user %s: %v", username, err)
	}
	// We need to use the actual user ID in logs, so update via the DB directly
	// to match the caller's expected userID
	_, err = db.Exec("UPDATE users SET id = ? WHERE id = ?", userID, u.ID)
	if err != nil {
		t.Fatalf("updating user id: %v", err)
	}
}

func TestFilteredLogsPagination(t *testing.T) {
	db := testutil.OpenTestDB(t)
	logs := store.NewRequestLogStore(db)

	insertTestServerAndUser(t, db, "srv-1", "server1", "user-1", "alice")

	// Create 5 logs
	for i := 0; i < 5; i++ {
		err := logs.Create(&store.RequestLog{
			Timestamp:    time.Now().Add(-time.Duration(i) * time.Minute),
			UserID:       "user-1",
			ServerID:     "srv-1",
			Method:       "tools/call",
			EndpointName: fmt.Sprintf("tool_%d", i),
			DurationMs:   100,
			Status:       "success",
		})
		if err != nil {
			t.Fatalf("Create() log %d error = %v", i, err)
		}
	}

	t.Run("first page", func(t *testing.T) {
		results, total, err := logs.FilteredLogs(store.LogFilter{Limit: 2, Offset: 0})
		if err != nil {
			t.Fatalf("FilteredLogs() error = %v", err)
		}
		if total != 5 {
			t.Errorf("total = %d, want 5", total)
		}
		if len(results) != 2 {
			t.Errorf("results count = %d, want 2", len(results))
		}
	})

	t.Run("second page", func(t *testing.T) {
		results, total, err := logs.FilteredLogs(store.LogFilter{Limit: 2, Offset: 2})
		if err != nil {
			t.Fatalf("FilteredLogs() error = %v", err)
		}
		if total != 5 {
			t.Errorf("total = %d, want 5", total)
		}
		if len(results) != 2 {
			t.Errorf("results count = %d, want 2", len(results))
		}
	})
}

func TestFilteredLogsByServer(t *testing.T) {
	db := testutil.OpenTestDB(t)
	logs := store.NewRequestLogStore(db)

	insertTestServerAndUser(t, db, "srv-1", "server1", "user-1", "alice")
	// Insert a second server
	if _, err := db.Exec(`INSERT INTO servers (id, name, display_name, server_type, config, status, created_at, updated_at)
		VALUES ('srv-2', 'server2', 'Server2', 'stdio', '{}', 'stopped', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}

	// 3 logs for srv-1, 2 for srv-2
	for i := 0; i < 3; i++ {
		if err := logs.Create(&store.RequestLog{
			Timestamp: time.Now(), UserID: "user-1", ServerID: "srv-1",
			Method: "tools/call", EndpointName: "a", DurationMs: 10, Status: "success",
		}); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := logs.Create(&store.RequestLog{
			Timestamp: time.Now(), UserID: "user-1", ServerID: "srv-2",
			Method: "tools/call", EndpointName: "b", DurationMs: 10, Status: "success",
		}); err != nil {
			t.Fatal(err)
		}
	}

	results, total, err := logs.FilteredLogs(store.LogFilter{ServerID: "srv-1", Limit: 50})
	if err != nil {
		t.Fatalf("FilteredLogs() error = %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(results) != 3 {
		t.Errorf("results count = %d, want 3", len(results))
	}
}

func TestFilteredLogsByUser(t *testing.T) {
	db := testutil.OpenTestDB(t)
	logs := store.NewRequestLogStore(db)

	insertTestServerAndUser(t, db, "srv-1", "server1", "user-1", "alice")
	// Insert a second user
	users := store.NewUserStore(db)
	u2, _ := users.Create("bob", "pass", "user")

	// 2 logs for user-1, 3 for user-2
	for i := 0; i < 2; i++ {
		if err := logs.Create(&store.RequestLog{
			Timestamp: time.Now(), UserID: "user-1", ServerID: "srv-1",
			Method: "tools/call", EndpointName: "a", DurationMs: 10, Status: "success",
		}); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 3; i++ {
		if err := logs.Create(&store.RequestLog{
			Timestamp: time.Now(), UserID: u2.ID, ServerID: "srv-1",
			Method: "tools/call", EndpointName: "b", DurationMs: 10, Status: "success",
		}); err != nil {
			t.Fatal(err)
		}
	}

	results, total, err := logs.FilteredLogs(store.LogFilter{UserID: "user-1", Limit: 50})
	if err != nil {
		t.Fatalf("FilteredLogs() error = %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(results) != 2 {
		t.Errorf("results count = %d, want 2", len(results))
	}
}

func TestFilteredLogsByStatus(t *testing.T) {
	db := testutil.OpenTestDB(t)
	logs := store.NewRequestLogStore(db)

	insertTestServerAndUser(t, db, "srv-1", "server1", "user-1", "alice")

	// 3 success, 2 error
	for i := 0; i < 3; i++ {
		if err := logs.Create(&store.RequestLog{
			Timestamp: time.Now(), UserID: "user-1", ServerID: "srv-1",
			Method: "tools/call", EndpointName: "a", DurationMs: 10, Status: "success",
		}); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := logs.Create(&store.RequestLog{
			Timestamp: time.Now(), UserID: "user-1", ServerID: "srv-1",
			Method: "tools/call", EndpointName: "b", DurationMs: 10, Status: "error", ErrorMsg: "fail",
		}); err != nil {
			t.Fatal(err)
		}
	}

	results, total, err := logs.FilteredLogs(store.LogFilter{Status: "error", Limit: 50})
	if err != nil {
		t.Fatalf("FilteredLogs() error = %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(results) != 2 {
		t.Errorf("results count = %d, want 2", len(results))
	}
}

func TestDistinctUsers(t *testing.T) {
	db := testutil.OpenTestDB(t)
	logs := store.NewRequestLogStore(db)

	insertTestServerAndUser(t, db, "srv-1", "server1", "user-1", "alice")
	users := store.NewUserStore(db)
	u2, _ := users.Create("bob", "pass", "user")

	if err := logs.Create(&store.RequestLog{
		Timestamp: time.Now(), UserID: "user-1", ServerID: "srv-1",
		Method: "tools/call", EndpointName: "a", DurationMs: 10, Status: "success",
	}); err != nil {
		t.Fatal(err)
	}
	if err := logs.Create(&store.RequestLog{
		Timestamp: time.Now(), UserID: u2.ID, ServerID: "srv-1",
		Method: "tools/call", EndpointName: "b", DurationMs: 10, Status: "success",
	}); err != nil {
		t.Fatal(err)
	}

	distinct, err := logs.DistinctUsers()
	if err != nil {
		t.Fatalf("DistinctUsers() error = %v", err)
	}
	if len(distinct) != 2 {
		t.Errorf("DistinctUsers() returned %d, want 2", len(distinct))
	}
}

func TestLogStats(t *testing.T) {
	db := testutil.OpenTestDB(t)
	logs := store.NewRequestLogStore(db)

	insertTestServerAndUser(t, db, "srv-1", "server1", "user-1", "alice")

	// Create 3 success and 1 error within the last 24h
	for i := 0; i < 3; i++ {
		if err := logs.Create(&store.RequestLog{
			Timestamp: time.Now(), UserID: "user-1", ServerID: "srv-1",
			Method: "tools/call", EndpointName: "a", DurationMs: 50, Status: "success",
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := logs.Create(&store.RequestLog{
		Timestamp: time.Now(), UserID: "user-1", ServerID: "srv-1",
		Method: "tools/call", EndpointName: "b", DurationMs: 50, Status: "error", ErrorMsg: "fail",
	}); err != nil {
		t.Fatal(err)
	}

	stats, err := logs.Stats()
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}
	if stats.TotalRequests24h != 4 {
		t.Errorf("TotalRequests24h = %d, want 4", stats.TotalRequests24h)
	}
	if stats.Errors24h != 1 {
		t.Errorf("Errors24h = %d, want 1", stats.Errors24h)
	}
}
