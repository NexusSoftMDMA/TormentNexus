package middleware_test

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/nacl/box"

	"github.com/comma-compliance/arc-relay/internal/middleware"
	"github.com/comma-compliance/arc-relay/internal/store"
	"github.com/comma-compliance/arc-relay/internal/testutil"
)

// noTickTicker returns a ticker factory that never fires (tests use Wake to trigger).
func noTickTicker() func(time.Duration) (<-chan time.Time, func()) {
	return func(d time.Duration) (<-chan time.Time, func()) {
		ch := make(chan time.Time)
		return ch, func() {}
	}
}

func newDispatcher(t *testing.T, handler http.Handler) (*middleware.ArchiveDispatcher, *store.ArchiveQueueStore, *httptest.Server) {
	t.Helper()
	db := testutil.OpenTestFileDB(t)
	qs := store.NewArchiveQueueStore(db, nil)
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	d := middleware.NewArchiveDispatcher(qs, nil)
	d.SetHTTPClient(ts.Client())
	d.NewTicker = noTickTicker()
	return d, qs, ts
}

func enqueueItem(t *testing.T, qs *store.ArchiveQueueStore, url string) {
	t.Helper()
	if err := qs.Enqueue(&store.ArchiveQueueItem{
		ServerID:     "srv-1",
		Payload:      `{"version":"v1","phase":"exchange"}`,
		URL:          url,
		AuthType:     "bearer",
		AuthValue:    "test-token",
		APIKeyHeader: "X-API-Key",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestDispatcher_SuccessfulDelivery(t *testing.T) {
	var received atomic.Int32
	var lastBody []byte
	var mu sync.Mutex

	d, qs, ts := newDispatcher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		body, _ := io.ReadAll(r.Body)
		lastBody = body
		mu.Unlock()
		received.Add(1)
		w.WriteHeader(200)
	}))
	d.Start()
	defer d.Stop()

	enqueueItem(t, qs, ts.URL)
	d.Wake()
	time.Sleep(200 * time.Millisecond)

	if received.Load() != 1 {
		t.Errorf("expected 1 delivery, got %d", received.Load())
	}
	mu.Lock()
	if string(lastBody) != `{"version":"v1","phase":"exchange"}` {
		t.Errorf("unexpected payload: %s", lastBody)
	}
	mu.Unlock()

	// Row should be deleted
	st, _ := qs.Status()
	if st.TotalCount != 0 {
		t.Errorf("expected empty queue after delivery, got %d", st.TotalCount)
	}
}

func TestDispatcher_VerifiesAuthHeader(t *testing.T) {
	var authHeader string
	var mu sync.Mutex

	d, qs, ts := newDispatcher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		authHeader = r.Header.Get("Authorization")
		mu.Unlock()
		w.WriteHeader(200)
	}))
	d.Start()
	defer d.Stop()

	enqueueItem(t, qs, ts.URL)
	d.Wake()
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if authHeader != "Bearer test-token" {
		t.Errorf("auth header = %s, want Bearer test-token", authHeader)
	}
	mu.Unlock()
}

func TestDispatcher_TransientFailure_Reschedules(t *testing.T) {
	d, qs, ts := newDispatcher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
	}))
	d.Start()
	defer d.Stop()

	enqueueItem(t, qs, ts.URL)
	d.Wake()
	time.Sleep(200 * time.Millisecond)

	st, _ := qs.Status()
	if st.PendingCount != 1 {
		t.Errorf("expected 1 pending, got %d", st.PendingCount)
	}
	if st.LastError != "HTTP 503" {
		t.Errorf("last_error = %s", st.LastError)
	}
}

func TestDispatcher_PermanentFailure_HoldsRow(t *testing.T) {
	d, qs, ts := newDispatcher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	}))
	d.Start()
	defer d.Stop()

	enqueueItem(t, qs, ts.URL)
	d.Wake()
	time.Sleep(200 * time.Millisecond)

	st, _ := qs.Status()
	if st.HoldCount != 1 {
		t.Errorf("expected 1 held, got %d", st.HoldCount)
	}
	if st.PendingCount != 0 {
		t.Errorf("expected 0 pending, got %d", st.PendingCount)
	}
}

func TestDispatcher_CircuitBreaker_OpensAfter5Failures(t *testing.T) {
	d, qs, ts := newDispatcher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	d.Start()
	defer d.Stop()

	for i := 0; i < 6; i++ {
		enqueueItem(t, qs, ts.URL)
	}
	d.Wake()
	time.Sleep(300 * time.Millisecond)

	status := d.Status()
	if status.CircuitState != "open" {
		t.Errorf("expected circuit open, got %s", status.CircuitState)
	}
}

func TestDispatcher_CircuitBreaker_ClosesOnSuccess(t *testing.T) {
	var failCount atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failCount.Add(1) <= 5 {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(200)
		}
	})

	db := testutil.OpenTestFileDB(t)
	qs := store.NewArchiveQueueStore(db, nil)
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	now := time.Now()
	var clockMu sync.Mutex
	currentTime := now

	d := middleware.NewArchiveDispatcher(qs, nil)
	d.SetHTTPClient(ts.Client())
	d.NewTicker = noTickTicker()
	d.NowFunc = func() time.Time {
		clockMu.Lock()
		defer clockMu.Unlock()
		return currentTime
	}
	d.Start()
	defer d.Stop()

	// 5 items to trigger circuit open
	for i := 0; i < 5; i++ {
		enqueueItem(t, qs, ts.URL)
	}
	d.Wake()
	time.Sleep(300 * time.Millisecond)

	if d.Status().CircuitState != "open" {
		t.Fatal("circuit should be open")
	}

	// Advance clock past 1min pause
	clockMu.Lock()
	currentTime = now.Add(2 * time.Minute)
	clockMu.Unlock()

	// Enqueue a new item that will succeed (probe)
	enqueueItem(t, qs, ts.URL)
	d.Wake()
	time.Sleep(300 * time.Millisecond)

	if d.Status().CircuitState != "closed" {
		t.Errorf("expected circuit closed, got %s", d.Status().CircuitState)
	}
}

func TestDispatcher_RetryHeld_ResetsAndDrains(t *testing.T) {
	var callCount atomic.Int32
	d, qs, ts := newDispatcher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callCount.Add(1) == 1 {
			w.WriteHeader(400) // first = permanent fail
		} else {
			w.WriteHeader(200) // retry = success
		}
	}))
	d.Start()
	defer d.Stop()

	enqueueItem(t, qs, ts.URL)
	d.Wake()
	time.Sleep(200 * time.Millisecond)

	st, _ := qs.Status()
	if st.HoldCount != 1 {
		t.Fatalf("expected 1 held, got %d", st.HoldCount)
	}

	count, err := d.RetryHeld()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("retried = %d", count)
	}
	time.Sleep(200 * time.Millisecond)

	st, _ = qs.Status()
	if st.TotalCount != 0 {
		t.Errorf("expected empty queue, got total=%d", st.TotalCount)
	}
}

func TestDispatcher_BackoffSchedule(t *testing.T) {
	now := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	d := middleware.NewArchiveDispatcher(nil, nil)
	d.NowFunc = func() time.Time { return now }

	tests := []struct {
		attempts int
		want     time.Duration
	}{
		{0, 15 * time.Second},
		{1, 1 * time.Minute},
		{2, 5 * time.Minute},
		{3, 15 * time.Minute},
		{4, 1 * time.Hour},
		{5, 6 * time.Hour},
		{100, 6 * time.Hour},
	}

	for _, tt := range tests {
		got := d.NextAttemptTime(tt.attempts)
		expected := now.Add(tt.want)
		if !got.Equal(expected) {
			t.Errorf("attempts=%d: got %v, want %v (diff %v)", tt.attempts, got, expected, got.Sub(expected))
		}
	}
}

func TestDispatcher_SendTest_DoesNotEnqueue(t *testing.T) {
	d, qs, ts := newDispatcher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["phase"] != "test" {
			t.Error("expected test phase payload")
		}
		w.WriteHeader(200)
	}))

	cfg := middleware.ArchiveConfig{
		URL:       ts.URL,
		AuthType:  "bearer",
		AuthValue: "test-token",
	}

	status, err := d.SendTest(cfg)
	if err != nil {
		t.Fatalf("SendTest: %v", err)
	}
	if status != 200 {
		t.Errorf("status = %d", status)
	}

	st, _ := qs.Status()
	if st.TotalCount != 0 {
		t.Errorf("SendTest should not enqueue, got %d items", st.TotalCount)
	}
}

func TestDispatcher_SendTest_ReportsFailure(t *testing.T) {
	d, _, ts := newDispatcher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))

	cfg := middleware.ArchiveConfig{
		URL:       ts.URL,
		AuthType:  "bearer",
		AuthValue: "bad-token",
	}

	status, err := d.SendTest(cfg)
	if err == nil {
		t.Error("expected error for 401")
	}
	if status != 401 {
		t.Errorf("status = %d, want 401", status)
	}
}

// TestDispatcher_SendTest_SealsWhenKeyConfigured verifies that when a
// recipient key is set, SendTest POSTs an envelope (not plaintext) to
// the ingest endpoint. This is load-bearing for the handoff flow:
// operators rely on the test delivery to catch a misconfigured key
// before the first real request.
func TestDispatcher_SendTest_SealsWhenKeyConfigured(t *testing.T) {
	pub, _, err := box.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate recipient keypair: %v", err)
	}
	b64Key := base64.StdEncoding.EncodeToString(pub[:])

	var received []byte
	d, _, ts := newDispatcher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, _ = io.ReadAll(r.Body)
		w.WriteHeader(200)
	}))

	cfg := middleware.ArchiveConfig{
		URL:              ts.URL,
		AuthType:         "bearer",
		AuthValue:        "test-token",
		NaClRecipientKey: b64Key,
	}

	if _, err := d.SendTest(cfg); err != nil {
		t.Fatalf("SendTest: %v", err)
	}

	var env struct {
		Version         string `json:"version"`
		KeyID           string `json:"kid"`
		Ciphertext      string `json:"ciphertext"`
		Nonce           string `json:"nonce"`
		SourcePublicKey string `json:"sourcePublicKey"`
	}
	if err := json.Unmarshal(received, &env); err != nil {
		t.Fatalf("received body is not a JSON envelope: %v (body: %s)", err, received)
	}
	if env.Version != middleware.EnvelopeVersion {
		t.Errorf("envelope.version = %q, want %q", env.Version, middleware.EnvelopeVersion)
	}
	if env.KeyID == "" {
		t.Error("envelope.kid is empty")
	}
	if env.Ciphertext == "" {
		t.Error("envelope.ciphertext is empty - SendTest sent plaintext?")
	}

	// Plaintext must not appear in the received body.
	if bytes.Contains(received, []byte("connectivity_test")) {
		t.Error("plaintext marker leaked into sealed body")
	}
}

// TestDispatcher_SendTest_RejectsInvalidKey verifies that a malformed
// recipient key surfaces as a SendTest error instead of being silently
// dropped or panicking the dispatcher.
func TestDispatcher_SendTest_RejectsInvalidKey(t *testing.T) {
	d, _, ts := newDispatcher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not have been reached - encryption must fail first")
		w.WriteHeader(200)
	}))

	cfg := middleware.ArchiveConfig{
		URL:              ts.URL,
		NaClRecipientKey: "not-a-valid-key!!!",
	}

	_, err := d.SendTest(cfg)
	if err == nil {
		t.Fatal("expected error for invalid nacl_recipient_key, got nil")
	}
}

func TestDispatcher_StartupReplay(t *testing.T) {
	var received atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received.Add(1)
		w.WriteHeader(200)
	})

	db := testutil.OpenTestFileDB(t)
	qs := store.NewArchiveQueueStore(db, nil)
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	// Pre-seed queue before dispatcher starts
	for i := 0; i < 3; i++ {
		if err := qs.Enqueue(&store.ArchiveQueueItem{
			ServerID:     "srv-1",
			Payload:      `{"pre":"seeded"}`,
			URL:          ts.URL,
			AuthType:     "none",
			APIKeyHeader: "X-API-Key",
		}); err != nil {
			t.Fatal(err)
		}
	}

	d := middleware.NewArchiveDispatcher(qs, nil)
	d.SetHTTPClient(ts.Client())
	d.NewTicker = noTickTicker()
	d.Start()
	defer d.Stop()

	time.Sleep(300 * time.Millisecond)

	if received.Load() != 3 {
		t.Errorf("expected 3 startup deliveries, got %d", received.Load())
	}
	st, _ := qs.Status()
	if st.TotalCount != 0 {
		t.Errorf("expected empty queue, got %d", st.TotalCount)
	}
}
