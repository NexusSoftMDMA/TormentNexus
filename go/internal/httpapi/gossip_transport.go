package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/gossip"
	"github.com/tormentnexushq/tormentnexus-go/internal/mesh"
)

type HTTPGossipTransport struct {
	mu      sync.Mutex
	localID string
	peers   *mesh.DiscoveryService
	onMsg   func(gossip.GossipMessage)
	client  *http.Client
}

func NewHTTPGossipTransport(localID string, peers *mesh.DiscoveryService) *HTTPGossipTransport {
	return &HTTPGossipTransport{
		localID: localID,
		peers:   peers,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (t *HTTPGossipTransport) Send(ctx context.Context, peerID string, msg gossip.GossipMessage) error {
	var targetAddr string
	for _, p := range t.peers.Peers() {
		if p.NodeID == peerID {
			targetAddr = fmt.Sprintf("http://%s:%d/api/gossip/message", p.Addr, p.Port)
			break
		}
	}
	if targetAddr == "" {
		return fmt.Errorf("peer %s address not found", peerID)
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", targetAddr, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (t *HTTPGossipTransport) Broadcast(ctx context.Context, msg gossip.GossipMessage) error {
	for _, p := range t.peers.Peers() {
		_ = t.Send(ctx, p.NodeID, msg)
	}
	return nil
}

func (t *HTTPGossipTransport) OnMessage(handler func(gossip.GossipMessage)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onMsg = handler
}

func (t *HTTPGossipTransport) Receive(msg gossip.GossipMessage) {
	t.mu.Lock()
	handler := t.onMsg
	t.mu.Unlock()
	if handler != nil {
		handler(msg)
	}
}
