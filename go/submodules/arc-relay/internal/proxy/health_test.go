package proxy

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/comma-compliance/arc-relay/internal/mcp"
	"github.com/comma-compliance/arc-relay/internal/store"
	"github.com/comma-compliance/arc-relay/internal/testutil"
)

func newTestHealthMonitor(t *testing.T) (*HealthMonitor, *store.ServerStore) {
	t.Helper()
	db := testutil.OpenTestDB(t)
	servers := store.NewServerStore(db, store.NewConfigEncryptor(""))
	mgr := &Manager{
		backends:   make(map[string]Backend),
		servers:    servers,
		containers: make(map[string]string),
		Endpoints:  mcp.NewEndpointCache(),
	}
	hm := NewHealthMonitor(mgr, servers, 30*time.Second)
	return hm, servers
}

func createTestServer(t *testing.T, servers *store.ServerStore, name string, stype store.ServerType) *store.Server {
	t.Helper()
	var cfg json.RawMessage
	switch stype {
	case store.ServerTypeStdio:
		cfg = json.RawMessage(`{"image":"test:latest"}`)
	case store.ServerTypeRemote:
		cfg = json.RawMessage(`{"url":"https://example.com/mcp"}`)
	case store.ServerTypeHTTP:
		cfg = json.RawMessage(`{"url":"https://example.com"}`)
	}
	srv := &store.Server{
		Name:       name,
		ServerType: stype,
		Config:     cfg,
	}
	if err := servers.Create(srv); err != nil {
		t.Fatalf("creating test server: %v", err)
	}
	_ = servers.UpdateStatus(srv.ID, store.StatusError, "test error")
	srv, _ = servers.Get(srv.ID)
	return srv
}

func TestIsStatelessServer(t *testing.T) {
	hm, servers := newTestHealthMonitor(t)

	tests := []struct {
		name     string
		slug     string
		stype    store.ServerType
		config   json.RawMessage
		expected bool
	}{
		{"remote is stateless", "test-remote", store.ServerTypeRemote, json.RawMessage(`{"url":"https://example.com/mcp"}`), true},
		{"external HTTP is stateless", "test-ext-http", store.ServerTypeHTTP, json.RawMessage(`{"url":"https://example.com"}`), true},
		{"Docker HTTP is not stateless", "test-docker-http", store.ServerTypeHTTP, json.RawMessage(`{}`), false},
		{"stdio is not stateless", "test-stdio", store.ServerTypeStdio, json.RawMessage(`{"image":"test:latest"}`), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &store.Server{
				Name:       tt.slug,
				ServerType: tt.stype,
				Config:     tt.config,
			}
			if err := servers.Create(srv); err != nil {
				t.Fatalf("creating server: %v", err)
			}
			got := hm.isStatelessServer(srv)
			if got != tt.expected {
				t.Errorf("isStatelessServer() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTryRecover_CooldownRespected(t *testing.T) {
	hm, servers := newTestHealthMonitor(t)
	srv := createTestServer(t, servers, "cooldown-test", store.ServerTypeRemote)

	// Pre-set state: last recovery was just now, 1 attempt used
	hm.mu.Lock()
	hm.lastRecover[srv.ID] = time.Now()
	hm.recoverAttempts[srv.ID] = 1
	hm.mu.Unlock()

	ctx := context.Background()
	hm.tryRecover(ctx, srv)

	// Should have been skipped due to cooldown
	hm.mu.Lock()
	attempts := hm.recoverAttempts[srv.ID]
	hm.mu.Unlock()
	if attempts != 1 {
		t.Errorf("after cooldown-blocked tryRecover, recoverAttempts = %d, want 1 (unchanged)", attempts)
	}
}

func TestTryRecover_MaxAttemptsExhausted(t *testing.T) {
	hm, servers := newTestHealthMonitor(t)
	srv := createTestServer(t, servers, "max-test", store.ServerTypeRemote)

	ctx := context.Background()

	// Simulate maxRecoverAttempts already reached
	hm.mu.Lock()
	hm.recoverAttempts[srv.ID] = maxRecoverAttempts
	hm.mu.Unlock()

	hm.tryRecover(ctx, srv)

	hm.mu.Lock()
	attempts := hm.recoverAttempts[srv.ID]
	hm.mu.Unlock()
	if attempts != maxRecoverAttempts {
		t.Errorf("after max attempts, recoverAttempts = %d, want %d (unchanged)", attempts, maxRecoverAttempts)
	}
}

func TestTryRecover_SkipsIfAlreadyRecovering(t *testing.T) {
	hm, servers := newTestHealthMonitor(t)
	srv := createTestServer(t, servers, "overlap-test", store.ServerTypeStdio)

	// Simulate an in-progress recovery
	hm.mu.Lock()
	hm.recovering[srv.ID] = true
	hm.mu.Unlock()

	ctx := context.Background()
	hm.tryRecover(ctx, srv)

	hm.mu.Lock()
	attempts := hm.recoverAttempts[srv.ID]
	hm.mu.Unlock()
	if attempts != 0 {
		t.Errorf("should skip while recovering, recoverAttempts = %d, want 0", attempts)
	}
}

func TestTryRecover_RemoteRecoveryIsInline(t *testing.T) {
	hm, servers := newTestHealthMonitor(t)
	srv := createTestServer(t, servers, "remote-inline", store.ServerTypeRemote)

	ctx := context.Background()
	hm.tryRecover(ctx, srv)

	// Remote recovery is inline (not async), so recovering flag should never be set
	hm.mu.Lock()
	recovering := hm.recovering[srv.ID]
	hm.mu.Unlock()
	if recovering {
		t.Error("remote server should recover inline, not async")
	}
}

func TestTryRecover_StdioUsesAsyncPath(t *testing.T) {
	hm, servers := newTestHealthMonitor(t)
	srv := createTestServer(t, servers, "stdio-async", store.ServerTypeStdio)

	// Verify that stdio servers are NOT identified as stateless
	// (which means they take the async Docker recovery path, not the inline path)
	if hm.isStatelessServer(srv) {
		t.Fatal("stdio server should not be stateless")
	}

	// Verify the attempt counter increments when tryRecover runs for stdio.
	// We can't actually call tryRecover here without a Docker manager (it would panic
	// in the background goroutine), but we can verify the branching logic via
	// isStatelessServer and the guard conditions.
	hm.mu.Lock()
	hm.recoverAttempts[srv.ID] = 0
	hm.mu.Unlock()

	// Simulate: if the server were already recovering, tryRecover should skip
	hm.mu.Lock()
	hm.recovering[srv.ID] = true
	hm.mu.Unlock()

	hm.tryRecover(context.Background(), srv)

	hm.mu.Lock()
	attempts := hm.recoverAttempts[srv.ID]
	hm.mu.Unlock()
	if attempts != 0 {
		t.Errorf("should skip while recovering, got attempts=%d", attempts)
	}
}

func TestResetRecoveryState(t *testing.T) {
	hm, _ := newTestHealthMonitor(t)
	serverID := "test-server-id"

	// Populate all tracking maps
	hm.mu.Lock()
	hm.failCounts[serverID] = 3
	hm.recoverAttempts[serverID] = 2
	hm.lastRecover[serverID] = time.Now()
	hm.mu.Unlock()

	hm.ResetRecoveryState(serverID)

	hm.mu.Lock()
	defer hm.mu.Unlock()

	if _, ok := hm.failCounts[serverID]; ok {
		t.Error("failCounts should be cleared")
	}
	if _, ok := hm.recoverAttempts[serverID]; ok {
		t.Error("recoverAttempts should be cleared")
	}
	if _, ok := hm.lastRecover[serverID]; ok {
		t.Error("lastRecover should be cleared")
	}
}

func TestResetFailCount_DoesNotResetRecoverAttempts(t *testing.T) {
	hm, _ := newTestHealthMonitor(t)
	serverID := "test-server-id"

	// Set both fail counts and recovery attempts
	hm.mu.Lock()
	hm.failCounts[serverID] = 3
	hm.recoverAttempts[serverID] = 2
	hm.mu.Unlock()

	// resetFailCount (called on healthy ping) should only clear failCounts
	hm.resetFailCount(serverID)

	hm.mu.Lock()
	defer hm.mu.Unlock()

	if _, ok := hm.failCounts[serverID]; ok {
		t.Error("failCounts should be cleared")
	}
	if attempts, ok := hm.recoverAttempts[serverID]; !ok || attempts != 2 {
		t.Errorf("recoverAttempts should be preserved, got %d (exists=%v)", attempts, ok)
	}
}

func TestTryRecover_SuccessfulRemoteResetsState(t *testing.T) {
	hm, servers := newTestHealthMonitor(t)
	srv := createTestServer(t, servers, "success-test", store.ServerTypeRemote)

	ctx := context.Background()
	hm.tryRecover(ctx, srv)

	// Remote recovery should succeed (startRemote works without Docker)
	// and ResetRecoveryState should clear everything
	hm.mu.Lock()
	attempts := hm.recoverAttempts[srv.ID]
	_, hasLastRecover := hm.lastRecover[srv.ID]
	hm.mu.Unlock()

	if attempts != 0 {
		t.Errorf("successful recovery should reset recoverAttempts, got %d", attempts)
	}
	if hasLastRecover {
		t.Error("successful recovery should clear lastRecover")
	}
}
