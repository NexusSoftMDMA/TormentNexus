package gossip

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

// ─── Mock Transport ──────────────────────────────────────────────────────────

type mockTransport struct {
	mu       sync.Mutex
	messages []GossipMessage
	handlers []func(msg GossipMessage)
}

func (t *mockTransport) Send(_ context.Context, _ string, msg GossipMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.messages = append(t.messages, msg)
	return nil
}

func (t *mockTransport) Broadcast(_ context.Context, msg GossipMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.messages = append(t.messages, msg)
	return nil
}

func (t *mockTransport) OnMessage(handler func(msg GossipMessage)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.handlers = append(t.handlers, handler)
}

func (t *mockTransport) Deliver(msg GossipMessage) {
	t.mu.Lock()
	handlers := make([]func(msg GossipMessage), len(t.handlers))
	copy(handlers, t.handlers)
	t.mu.Unlock()

	for _, h := range handlers {
		h(msg)
	}
}

func (t *mockTransport) GetMessages() []GossipMessage {
	t.mu.Lock()
	defer t.mu.Unlock()
	result := make([]GossipMessage, len(t.messages))
	copy(result, t.messages)
	return result
}

// ─── Mock State Store ────────────────────────────────────────────────────────

type mockStore struct {
	mu      sync.RWMutex
	entries map[string]StateEntry
	clock   VectorClock
	nodeID  string
}

func newMockStore(nodeID string) *mockStore {
	return &mockStore{
		entries: make(map[string]StateEntry),
		clock:   VectorClock{nodeID: 0},
		nodeID:  nodeID,
	}
}

func (s *mockStore) GetDigest(_ context.Context) ([]DigestEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]DigestEntry, 0, len(s.entries))
	for _, e := range s.entries {
		result = append(result, DigestEntry{ID: e.ID, Version: e.Version, Hash: e.Hash})
	}
	return result, nil
}

func (s *mockStore) GetEntries(_ context.Context, ids []string) ([]StateEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]StateEntry, 0, len(ids))
	for _, id := range ids {
		if e, ok := s.entries[id]; ok {
			result = append(result, e)
		}
	}
	return result, nil
}

func (s *mockStore) Merge(_ context.Context, entries []StateEntry) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	accepted := 0
	for _, e := range entries {
		existing, ok := s.entries[e.ID]
		if !ok || e.Version > existing.Version {
			s.entries[e.ID] = e
			accepted++
		}
	}
	return accepted, nil
}

func (s *mockStore) LocalClock(_ context.Context) (VectorClock, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(VectorClock, len(s.clock))
	for k, v := range s.clock {
		result[k] = v
	}
	return result, nil
}

func (s *mockStore) IncrementClock(_ context.Context) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clock[s.nodeID]++
	return s.clock[s.nodeID], nil
}

func (s *mockStore) AddEntry(entry StateEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[entry.ID] = entry
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestProtocolCreation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NodeID = "test-node"
	transport := &mockTransport{}
	store := newMockStore("test-node")

	proto, err := NewProtocol(cfg, transport, store)
	if err != nil {
		t.Fatalf("NewProtocol failed: %v", err)
	}
	if proto.cfg.NodeID != "test-node" {
		t.Errorf("NodeID mismatch: got %q", proto.cfg.NodeID)
	}
}

func TestPeerManagement(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NodeID = "test-node"
	transport := &mockTransport{}
	store := newMockStore("test-node")

	proto, _ := NewProtocol(cfg, transport, store)

	proto.AddPeer("peer-1")
	proto.AddPeer("peer-2")
	proto.AddPeer("peer-3")

	if proto.PeerCount() != 3 {
		t.Errorf("Expected 3 peers, got %d", proto.PeerCount())
	}

	proto.RemovePeer("peer-2")
	if proto.PeerCount() != 2 {
		t.Errorf("Expected 2 peers after removal, got %d", proto.PeerCount())
	}
}

func TestSyncRequestHandling(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NodeID = "node-a"
	transport := &mockTransport{}
	store := newMockStore("node-a")

	// Add local entry that peer doesn't have
	store.AddEntry(StateEntry{
		ID:      "entry-1",
		Type:    "working",
		Content: "local memory",
		Version: 5,
		Hash:    "abc123",
		Origin:  "node-a",
	})

	proto, _ := NewProtocol(cfg, transport, store)

	// Simulate a sync request from peer
	reqPayload, _ := json.Marshal(SyncRequestPayload{
		Digests: []DigestEntry{
			{ID: "entry-2", Version: 3, Hash: "def456"}, // We don't have this
		},
		Clock: VectorClock{"node-b": 2},
	})

	msg := GossipMessage{
		Type:      TypeSyncRequest,
		SenderID:  "node-b",
		Timestamp: time.Now().UnixMilli(),
		Payload:   reqPayload,
	}

	proto.handleMessage(msg)

	// Should have sent a sync response and a state update
	messages := transport.GetMessages()
	if len(messages) < 1 {
		t.Fatalf("Expected at least 1 message, got %d", len(messages))
	}

	// Check that a state update was sent with entry-1
	found := false
	for _, m := range messages {
		if m.Type == TypeStateUpdate {
			var update StateUpdatePayload
			if err := json.Unmarshal(m.Payload, &update); err == nil {
				for _, e := range update.Entries {
					if e.ID == "entry-1" {
						found = true
						break
					}
				}
			}
		}
	}
	if !found {
		t.Error("Expected state update with entry-1 to be sent to peer")
	}
}

func TestStateUpdateHandling(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NodeID = "node-a"
	transport := &mockTransport{}
	store := newMockStore("node-a")

	proto, _ := NewProtocol(cfg, transport, store)

	// Receive a state update from peer
	updatePayload, _ := json.Marshal(StateUpdatePayload{
		Entries: []StateEntry{
			{
				ID:      "entry-from-b",
				Type:    "working",
				Content: "peer memory",
				Version: 3,
				Hash:    "xyz789",
				Origin:  "node-b",
			},
		},
	})

	msg := GossipMessage{
		Type:      TypeStateUpdate,
		SenderID:  "node-b",
		Timestamp: time.Now().UnixMilli(),
		Payload:   updatePayload,
	}

	proto.handleMessage(msg)

	// Entry should be merged into our store
	entries, _ := store.GetEntries(context.Background(), []string{"entry-from-b"})
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Content != "peer memory" {
		t.Errorf("Content mismatch: got %q", entries[0].Content)
	}
}

func TestLastWriteWinsMerge(t *testing.T) {
	store := newMockStore("node-a")

	// Add initial entry
	store.AddEntry(StateEntry{
		ID:      "entry-1",
		Type:    "working",
		Content: "original",
		Version: 1,
		Origin:  "node-a",
	})

	// Merge with higher version
	accepted, err := store.Merge(context.Background(), []StateEntry{
		{
			ID:      "entry-1",
			Type:    "working",
			Content: "updated by peer",
			Version: 3,
			Origin:  "node-b",
		},
	})

	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}
	if accepted != 1 {
		t.Errorf("Expected 1 accepted, got %d", accepted)
	}

	entries, _ := store.GetEntries(context.Background(), []string{"entry-1"})
	if entries[0].Content != "updated by peer" {
		t.Errorf("Should have updated to newer version, got %q", entries[0].Content)
	}

	// Merge with lower version should be rejected
	accepted, _ = store.Merge(context.Background(), []StateEntry{
		{
			ID:      "entry-1",
			Type:    "working",
			Content: "stale update",
			Version: 2,
			Origin:  "node-c",
		},
	})

	if accepted != 0 {
		t.Errorf("Expected 0 accepted for stale version, got %d", accepted)
	}
}

func TestBroadcastUpdate(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NodeID = "node-a"
	transport := &mockTransport{}
	store := newMockStore("node-a")

	proto, _ := NewProtocol(cfg, transport, store)

	err := proto.BroadcastUpdate(context.Background(), []StateEntry{
		{
			ID:      "broadcast-entry",
			Type:    "long_term",
			Content: "broadcast content",
			Version: 1,
			Origin:  "node-a",
		},
	})

	if err != nil {
		t.Fatalf("BroadcastUpdate failed: %v", err)
	}

	messages := transport.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}
	if messages[0].Type != TypeStateUpdate {
		t.Errorf("Expected TypeStateUpdate, got %d", messages[0].Type)
	}
	if len(messages[0].Signature) == 0 {
		t.Error("Message should be signed")
	}
}

func TestPingPong(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NodeID = "node-a"
	transport := &mockTransport{}
	store := newMockStore("node-a")

	proto, _ := NewProtocol(cfg, transport, store)

	ping := GossipMessage{
		Type:      TypePing,
		SenderID:  "node-b",
		Timestamp: time.Now().UnixMilli(),
		Payload:   json.RawMessage(`{}`),
	}

	proto.handleMessage(ping)

	// Should respond with pong
	messages := transport.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 pong message, got %d", len(messages))
	}
	if messages[0].Type != TypePong {
		t.Errorf("Expected TypePong, got %d", messages[0].Type)
	}

	// Peer should be added
	if proto.PeerCount() != 1 {
		t.Errorf("Expected 1 peer, got %d", proto.PeerCount())
	}
}

func TestPeerSetEviction(t *testing.T) {
	ps := NewPeerSet()

	ps.Add("fresh")
	ps.peers["stale"] = time.Now().Add(-2 * time.Hour) // Manually set old time

	evicted := ps.EvictStale(time.Now().Add(-1 * time.Hour))
	if evicted != 1 {
		t.Errorf("Expected 1 evicted, got %d", evicted)
	}
	if ps.Size() != 1 {
		t.Errorf("Expected 1 remaining peer, got %d", ps.Size())
	}
}

func TestVectorClock(t *testing.T) {
	store := newMockStore("node-a")

	v, err := store.IncrementClock(context.Background())
	if err != nil {
		t.Fatalf("IncrementClock failed: %v", err)
	}
	if v != 1 {
		t.Errorf("Expected clock value 1, got %d", v)
	}

	v, _ = store.IncrementClock(context.Background())
	if v != 2 {
		t.Errorf("Expected clock value 2, got %d", v)
	}

	clock, _ := store.LocalClock(context.Background())
	if clock["node-a"] != 2 {
		t.Errorf("Expected clock['node-a'] = 2, got %d", clock["node-a"])
	}
}

func TestMessageSerialization(t *testing.T) {
	msg := GossipMessage{
		Type:      TypeSyncRequest,
		SenderID:  "node-a",
		Timestamp: time.Now().UnixMilli(),
		Payload:   json.RawMessage(`{"digests":[]}`),
		Signature: []byte{1, 2, 3, 4},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded GossipMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Type != TypeSyncRequest {
		t.Errorf("Type mismatch: got %d", decoded.Type)
	}
	if decoded.SenderID != "node-a" {
		t.Errorf("SenderID mismatch: got %q", decoded.SenderID)
	}
}

func TestProtocolStartStop(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NodeID = "test-node"
	cfg.GossipInterval = 100 * time.Millisecond
	transport := &mockTransport{}
	store := newMockStore("test-node")

	proto, _ := NewProtocol(cfg, transport, store)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	proto.Start(ctx)

	// Let it run for a moment
	time.Sleep(200 * time.Millisecond)

	proto.Stop()
}

func TestPeerSignatureVerification(t *testing.T) {
	cfgA := DefaultConfig()
	cfgA.NodeID = "node-a"
	storeA := newMockStore("node-a")
	protoA, _ := NewProtocol(cfgA, &mockTransport{}, storeA)

	cfgB := DefaultConfig()
	cfgB.NodeID = "node-b"
	storeB := newMockStore("node-b")
	protoB, _ := NewProtocol(cfgB, &mockTransport{}, storeB)

	// A learns B's key
	protoA.AddPeerKey("node-b", protoB.GetPublicKey())

	// B signs a message
	msg := GossipMessage{
		Type:      TypePing,
		SenderID:  "node-b",
		Timestamp: time.Now().UnixMilli(),
		Payload:   json.RawMessage(`{}`),
	}
	sig, _ := protoB.signMessage(msg)
	msg.Signature = sig

	// A verifies message
	if !protoA.verifyMessage(msg) {
		t.Error("Expected A to successfully verify signed message from B")
	}

	// Change payload to simulate tampering
	msg.Payload = json.RawMessage(`{"malicious":true}`)
	if protoA.verifyMessage(msg) {
		t.Error("Expected verification to fail for tampered payload")
	}
}

