package orchestrator

import (
	"sync"
	"testing"

	"github.com/gofiber/websocket/v2"
)

func TestTelemetrySocketInitialState(t *testing.T) {
	svc := NewTelemetrySocket()

	if svc.clients == nil {
		t.Fatal("Telemetry clients map failed initialization mapping.")
	}

	// Simulate map append loop internally without network boundaries
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		svc.mu.Lock()
		svc.clients[&websocket.Conn{}] = true
		svc.mu.Unlock()
	}()

	wg.Wait()

	if len(svc.clients) != 1 {
		t.Fatalf("Telemetry connection slice improperly sized: %d", len(svc.clients))
	}
}
