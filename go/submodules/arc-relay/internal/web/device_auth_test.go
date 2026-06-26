package web

import (
	"testing"
	"time"
)

// newTestDeviceAuthStore creates a deviceAuthStore without the background
// cleanup goroutine, suitable for deterministic tests.
func newTestDeviceAuthStore() *deviceAuthStore {
	return &deviceAuthStore{
		requests: make(map[string]*deviceAuthRequest),
		byUser:   make(map[string]string),
	}
}

func TestDeviceAuthStore_CreateAndGet(t *testing.T) {
	s := newTestDeviceAuthStore()

	req, err := s.create()
	if err != nil {
		t.Fatalf("create() returned error: %v", err)
	}
	if req == nil {
		t.Fatal("create() returned nil")
	}
	if req.DeviceCode == "" {
		t.Fatal("DeviceCode is empty")
	}
	if req.UserCode == "" {
		t.Fatal("UserCode is empty")
	}
	if req.Status != "pending" {
		t.Fatalf("Status = %q, want %q", req.Status, "pending")
	}

	// get by device code
	got := s.get(req.DeviceCode)
	if got == nil {
		t.Fatal("get(deviceCode) returned nil")
	}
	if got.DeviceCode != req.DeviceCode {
		t.Errorf("get() DeviceCode = %q, want %q", got.DeviceCode, req.DeviceCode)
	}

	// get by user code
	gotByUser := s.getByUserCode(req.UserCode)
	if gotByUser == nil {
		t.Fatal("getByUserCode() returned nil")
	}
	if gotByUser.DeviceCode != req.DeviceCode {
		t.Errorf("getByUserCode() DeviceCode = %q, want %q", gotByUser.DeviceCode, req.DeviceCode)
	}
}

func TestDeviceAuthStore_ApproveAndConsume(t *testing.T) {
	s := newTestDeviceAuthStore()

	req, err := s.create()
	if err != nil {
		t.Fatalf("create() returned error: %v", err)
	}
	s.approve(req.DeviceCode, "test-key")

	got := s.consume(req.DeviceCode)
	if got == nil {
		t.Fatal("consume() returned nil after approve")
	}
	if got.Status != "approved" {
		t.Errorf("Status = %q, want %q", got.Status, "approved")
	}
	if got.APIKey != "test-key" {
		t.Errorf("APIKey = %q, want %q", got.APIKey, "test-key")
	}
}

func TestDeviceAuthStore_DenyAndConsume(t *testing.T) {
	s := newTestDeviceAuthStore()

	req, err := s.create()
	if err != nil {
		t.Fatalf("create() returned error: %v", err)
	}
	s.deny(req.DeviceCode)

	got := s.consume(req.DeviceCode)
	if got == nil {
		t.Fatal("consume() returned nil after deny")
	}
	if got.Status != "denied" {
		t.Errorf("Status = %q, want %q", got.Status, "denied")
	}
}

func TestDeviceAuthStore_ExpiredRequest(t *testing.T) {
	s := newTestDeviceAuthStore()

	req, err := s.create()
	if err != nil {
		t.Fatalf("create() returned error: %v", err)
	}

	// Manually expire the request
	s.mu.Lock()
	s.requests[req.DeviceCode].ExpiresAt = time.Now().Add(-1 * time.Minute)
	s.mu.Unlock()

	got := s.get(req.DeviceCode)
	if got != nil {
		t.Errorf("get() should return nil for expired request, got %+v", got)
	}
}

func TestDeviceAuthStore_DoubleConsume(t *testing.T) {
	s := newTestDeviceAuthStore()

	req, err := s.create()
	if err != nil {
		t.Fatalf("create() returned error: %v", err)
	}
	s.approve(req.DeviceCode, "test-key")

	first := s.consume(req.DeviceCode)
	if first == nil {
		t.Fatal("first consume() returned nil")
	}
	if first.Status != "approved" {
		t.Errorf("first consume Status = %q, want %q", first.Status, "approved")
	}

	second := s.consume(req.DeviceCode)
	if second != nil {
		t.Errorf("second consume() should return nil, got %+v", second)
	}
}

func TestDeviceAuthStore_ConsumePendingNotRemoved(t *testing.T) {
	s := newTestDeviceAuthStore()

	req, err := s.create()
	if err != nil {
		t.Fatalf("create() returned error: %v", err)
	}

	// Consume without approve/deny - should return pending request without removing it
	got := s.consume(req.DeviceCode)
	if got == nil {
		t.Fatal("consume() returned nil for pending request")
	}
	if got.Status != "pending" {
		t.Errorf("Status = %q, want %q", got.Status, "pending")
	}

	// Request should still be in the store
	stillThere := s.get(req.DeviceCode)
	if stillThere == nil {
		t.Fatal("get() returned nil - pending request was incorrectly removed from store")
	}
}
