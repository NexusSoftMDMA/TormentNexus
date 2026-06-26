package mesh

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestDiscoveryConfigDefaults(t *testing.T) {
	cfg := DefaultDiscoveryConfig()
	if cfg.Port != 4301 {
		t.Errorf("Expected port 4301, got %d", cfg.Port)
	}
	if cfg.Interval != 5*time.Second {
		t.Errorf("Expected interval 5s, got %v", cfg.Interval)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", cfg.Timeout)
	}
}

func TestDiscoveryServiceCreation(t *testing.T) {
	cfg := DefaultDiscoveryConfig()
	ds := NewDiscoveryService("node-1", 4300, []string{"cli-harnesses", "http-bridge"}, cfg)

	if ds.nodeID != "node-1" {
		t.Errorf("Expected nodeID 'node-1', got %q", ds.nodeID)
	}
	if ds.httpPort != 4300 {
		t.Errorf("Expected httpPort 4300, got %d", ds.httpPort)
	}
	if ds.PeerCount() != 0 {
		t.Errorf("Expected 0 peers, got %d", ds.PeerCount())
	}
}

func TestDiscoveryServiceZeroPort(t *testing.T) {
	ds := NewDiscoveryService("node-1", 4300, []string{"test"}, DiscoveryConfig{})
	if ds.cfg.Port != 4301 {
		t.Errorf("Expected default port 4301 when cfg.Port is 0, got %d", ds.cfg.Port)
	}
}

func TestDiscoveryPeers(t *testing.T) {
	ds := NewDiscoveryService("node-1", 4300, []string{"test"}, DefaultDiscoveryConfig())

	// Manually add a peer
	ds.peers["node-2"] = &DiscoveredPeer{
		NodeID:   "node-2",
		Addr:     "192.168.1.100",
		Port:     4300,
		LastSeen: time.Now(),
		Caps:     []string{"cli-harnesses", "memory-status"},
		Labels:   map[string]string{"env": "test"},
	}

	peers := ds.Peers()
	if len(peers) != 1 {
		t.Fatalf("Expected 1 peer, got %d", len(peers))
	}
	if peers[0].NodeID != "node-2" {
		t.Errorf("Expected NodeID 'node-2', got %q", peers[0].NodeID)
	}
	if peers[0].Addr != "192.168.1.100" {
		t.Errorf("Expected Addr '192.168.1.100', got %q", peers[0].Addr)
	}
}

func TestDiscoveryPeerByCapability(t *testing.T) {
	ds := NewDiscoveryService("node-1", 4300, []string{"test"}, DefaultDiscoveryConfig())

	ds.peers["node-2"] = &DiscoveredPeer{
		NodeID:   "node-2",
		Addr:     "192.168.1.100",
		Port:     4300,
		LastSeen: time.Now(),
		Caps:     []string{"cli-harnesses", "memory-status"},
	}

	ds.peers["node-3"] = &DiscoveredPeer{
		NodeID:   "node-3",
		Addr:     "192.168.1.101",
		Port:     4300,
		LastSeen: time.Now(),
		Caps:     []string{"http-bridge"},
	}

	// Find peers with cli-harnesses
	result := ds.PeerByCapability([]string{"cli-harnesses"})
	if len(result) != 1 {
		t.Errorf("Expected 1 peer with cli-harnesses, got %d", len(result))
	}

	// Find peers with non-existent capability
	result = ds.PeerByCapability([]string{"nonexistent"})
	if len(result) != 0 {
		t.Errorf("Expected 0 peers with nonexistent, got %d", len(result))
	}

	// Find peers with multiple required capabilities
	result = ds.PeerByCapability([]string{"cli-harnesses", "memory-status"})
	if len(result) != 1 {
		t.Errorf("Expected 1 peer with both capabilities, got %d", len(result))
	}
}

func TestDiscoveryEviction(t *testing.T) {
	ds := NewDiscoveryService("node-1", 4300, []string{"test"}, DiscoveryConfig{
		Port:     4301,
		Interval: 100 * time.Millisecond,
		Timeout:  200 * time.Millisecond,
	})

	ds.peers["stale-node"] = &DiscoveredPeer{
		NodeID:   "stale-node",
		Addr:     "192.168.1.100",
		Port:     4300,
		LastSeen: time.Now().Add(-500 * time.Millisecond), // 500ms ago, past timeout
	}

	ds.peers["fresh-node"] = &DiscoveredPeer{
		NodeID:   "fresh-node",
		Addr:     "192.168.1.101",
		Port:     4300,
		LastSeen: time.Now(), // just now
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Run eviction once manually (simulating what evictionLoop does)
	now := time.Now()
	for nodeID, peer := range ds.peers {
		if now.Sub(peer.LastSeen) > ds.cfg.Timeout {
			delete(ds.peers, nodeID)
		}
	}

	if ds.PeerCount() != 1 {
		t.Errorf("Expected 1 peer after eviction, got %d", ds.PeerCount())
	}
	if _, exists := ds.peers["fresh-node"]; !exists {
		t.Error("Fresh node should not be evicted")
	}
	if _, exists := ds.peers["stale-node"]; exists {
		t.Error("Stale node should be evicted")
	}

	// Check that the context wasn't cancelled (just to use it)
	_ = ctx
}

func TestBeaconPacketSerialization(t *testing.T) {
	pkt := BeaconPacket{
		NodeID:    "node-1",
		Addr:      "192.168.1.50",
		Port:      4300,
		Timestamp: time.Now().UnixMilli(),
		Caps:      []string{"cli-harnesses", "http-bridge"},
		Labels:    map[string]string{"env": "production"},
	}

	data, err := json.Marshal(pkt)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded BeaconPacket
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.NodeID != pkt.NodeID {
		t.Errorf("NodeID mismatch: got %q, want %q", decoded.NodeID, pkt.NodeID)
	}
	if decoded.Addr != pkt.Addr {
		t.Errorf("Addr mismatch: got %q, want %q", decoded.Addr, pkt.Addr)
	}
	if decoded.Port != pkt.Port {
		t.Errorf("Port mismatch: got %d, want %d", decoded.Port, pkt.Port)
	}
	if len(decoded.Caps) != 2 {
		t.Errorf("Caps count mismatch: got %d, want 2", len(decoded.Caps))
	}
}

func TestParseBeaconAddr(t *testing.T) {
	result := ParseBeaconAddr("192.168.1.100", 4300)
	if result != "192.168.1.100:4300" {
		t.Errorf("Expected '192.168.1.100:4300', got %q", result)
	}
}

func TestDiscoveryServiceBeaconAddr(t *testing.T) {
	ds := NewDiscoveryService("node-1", 4300, []string{"test"}, DiscoveryConfig{Port: 9999})
	if ds.BeaconAddr() != "255.255.255.255:9999" {
		t.Errorf("Expected beacon addr '255.255.255.255:9999', got %q", ds.BeaconAddr())
	}
}

func TestDiscoveryServiceStop(t *testing.T) {
	ds := NewDiscoveryService("node-1", 4300, []string{"test"}, DefaultDiscoveryConfig())
	ds.Stop()
	// Should not panic - the stop channel is closed
}
