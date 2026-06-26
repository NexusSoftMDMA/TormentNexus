package orchestrator

import (
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// TelemetrySocket pushes hyper-fast thought process/token updates to Next.js clients
type TelemetrySocket struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

func NewTelemetrySocket() *TelemetrySocket {
	return &TelemetrySocket{
		clients: make(map[*websocket.Conn]bool),
	}
}

// WsHandler mounts directly into Fiber's router.
func (t *TelemetrySocket) WsHandler(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// ConnectionLoop captures live streams exactly like Hono/SST architectures.
func (t *TelemetrySocket) ConnectionLoop(c *websocket.Conn) {
	t.mu.Lock()
	t.clients[c] = true
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		delete(t.clients, c)
		t.mu.Unlock()
		c.Close()
	}()

	log.Printf("[Websocket] Dashboard attached. Live LLM tokens intercepting.")

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}
		log.Printf("[Websocket] Front-end packet: %s", msg)
	}
}

func (t *TelemetrySocket) Broadcast(payload string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for client := range t.clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(payload))
		if err != nil {
			client.Close()
			delete(t.clients, client)
		}
	}
}
