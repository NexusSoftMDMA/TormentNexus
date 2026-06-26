package proxy

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/comma-compliance/arc-relay/internal/mcp"
)

// StdioBridge translates between HTTP requests and a stdio MCP server process.
// It serializes requests (one at a time) and matches responses by JSON-RPC ID.
type StdioBridge struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser

	mu        sync.Mutex
	pending   map[string]chan *mcp.Response
	scanner   *bufio.Scanner
	closed    bool
	closeOnce sync.Once
}

// NewStdioBridge creates a bridge from an attached container's stdin/stdout.
func NewStdioBridge(stdin io.WriteCloser, stdout io.ReadCloser) *StdioBridge {
	b := &StdioBridge{
		stdin:   stdin,
		stdout:  stdout,
		pending: make(map[string]chan *mcp.Response),
		scanner: bufio.NewScanner(stdout),
	}
	// Increase scanner buffer for large responses
	b.scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	go b.readLoop()
	return b
}

// readLoop continuously reads newline-delimited JSON-RPC responses from stdout.
func (b *StdioBridge) readLoop() {
	for b.scanner.Scan() {
		line := b.scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var resp mcp.Response
		if err := json.Unmarshal(line, &resp); err != nil {
			slog.Warn("stdio bridge: failed to parse response", "err", err)
			continue
		}

		// Match response to pending request by ID
		if resp.ID != nil {
			idKey := string(resp.ID)
			b.mu.Lock()
			ch, ok := b.pending[idKey]
			if ok {
				delete(b.pending, idKey)
			}
			b.mu.Unlock()

			if ok {
				ch <- &resp
			} else {
				slog.Warn("stdio bridge: no pending request for response", "id", idKey)
			}
		}
	}
	if err := b.scanner.Err(); err != nil {
		slog.Warn("stdio bridge: read error", "err", err)
	}
}

// Send sends an MCP request over stdio and waits for the response.
func (b *StdioBridge) Send(ctx context.Context, req *mcp.Request) (*mcp.Response, error) {
	if b.closed {
		return nil, fmt.Errorf("bridge is closed")
	}

	idKey := string(req.ID)
	ch := make(chan *mcp.Response, 1)

	b.mu.Lock()
	b.pending[idKey] = ch
	b.mu.Unlock()

	// Serialize to JSON + newline
	data, err := json.Marshal(req)
	if err != nil {
		b.mu.Lock()
		delete(b.pending, idKey)
		b.mu.Unlock()
		return nil, fmt.Errorf("marshaling request: %w", err)
	}
	data = append(data, '\n')

	// Write to stdin (serialized — one request at a time for stdio)
	if _, err := b.stdin.Write(data); err != nil {
		b.mu.Lock()
		delete(b.pending, idKey)
		b.mu.Unlock()
		return nil, fmt.Errorf("writing to stdin: %w", err)
	}

	// Wait for response or context cancellation
	select {
	case resp := <-ch:
		return resp, nil
	case <-ctx.Done():
		b.mu.Lock()
		delete(b.pending, idKey)
		b.mu.Unlock()
		return nil, ctx.Err()
	}
}

// SendNotification sends a notification (no response expected).
func (b *StdioBridge) SendNotification(notification *mcp.Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("marshaling notification: %w", err)
	}
	data = append(data, '\n')
	_, err = b.stdin.Write(data)
	return err
}

// Close shuts down the bridge.
func (b *StdioBridge) Close() error {
	b.closeOnce.Do(func() {
		b.closed = true
		_ = b.stdin.Close()
		_ = b.stdout.Close()

		// Drain any pending requests
		b.mu.Lock()
		for id, ch := range b.pending {
			close(ch)
			delete(b.pending, id)
		}
		b.mu.Unlock()
	})
	return nil
}
