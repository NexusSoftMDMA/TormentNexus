package orchestration

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type A2AMessageType string

const (
	TaskRequest      A2AMessageType = "TASK_REQUEST"
	TaskResponse     A2AMessageType = "TASK_RESPONSE"
	TaskNegotiation  A2AMessageType = "TASK_NEGOTIATION"
	CapabilityReport A2AMessageType = "CAPABILITY_REPORT"
	ConsensusVoteMsg A2AMessageType = "CONSENSUS_VOTE"
	StateUpdate      A2AMessageType = "STATE_UPDATE"
	Handoff          A2AMessageType = "HANDOFF"
	DebateProposal   A2AMessageType = "DEBATE_PROPOSAL"
	Critique         A2AMessageType = "CRITIQUE"
	Heartbeat        A2AMessageType = "HEARTBEAT"
)

type A2AMessage struct {
	ID        string         `json:"id"`
	Timestamp int64          `json:"timestamp"`
	Sender    string         `json:"sender"`
	Recipient string         `json:"recipient,omitempty"`
	Type      A2AMessageType `json:"type"`
	Payload   interface{}    `json:"payload"`
	ReplyTo   string         `json:"replyTo,omitempty"`
}

type A2ABroker struct {
	mu               sync.RWMutex
	agents           map[string]chan A2AMessage
	heartbeats       map[string]int64
	history          []A2AMessage
	pendingResponses map[string]chan A2AMessage
	logger           *A2ALogger
	signalProcessor  FleetSignalProcessor
	bus              interface {
		EmitEvent(eventType string, source string, payload interface{})
	}
}

func NewA2ABroker(logger *A2ALogger) *A2ABroker {
	b := &A2ABroker{
		agents:           make(map[string]chan A2AMessage),
		heartbeats:       make(map[string]int64),
		history:          make([]A2AMessage, 0),
		pendingResponses: make(map[string]chan A2AMessage),
		logger:           logger,
	}
	go b.startHeartbeatMonitor()
	return b
}

func (b *A2ABroker) SetSignalProcessor(proc FleetSignalProcessor) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.signalProcessor = proc
}

func (b *A2ABroker) Query(ctx context.Context, msg A2AMessage) (A2AMessage, error) {
	b.mu.Lock()
	ch := make(chan A2AMessage, 1)
	b.pendingResponses[msg.ID] = ch
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.pendingResponses, msg.ID)
		b.mu.Unlock()
	}()

	b.RouteMessage(msg)

	select {
	case <-ctx.Done():
		return A2AMessage{}, ctx.Err()
	case resp := <-ch:
		return resp, nil
	}
}

func (b *A2ABroker) startHeartbeatMonitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now().UnixMilli()
		b.mu.Lock()
		for id, lastSeen := range b.heartbeats {
			if now-lastSeen > 30000 {
				log.Printf("[Go A2A] Agent %s timed out. Pruned.", id)
				if ch, ok := b.agents[id]; ok {
					close(ch)
					delete(b.agents, id)
				}
				delete(b.heartbeats, id)
			}
		}
		b.mu.Unlock()
	}
}

func (b *A2ABroker) RegisterAgent(id string) chan A2AMessage {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan A2AMessage, 100)
	b.agents[id] = ch
	b.heartbeats[id] = time.Now().UnixMilli()

	go b.RouteMessage(A2AMessage{
		ID:        fmt.Sprintf("cap-req-%s-%d", id, time.Now().UnixMilli()),
		Timestamp: time.Now().UnixMilli(),
		Sender:    "BROKER",
		Recipient: id,
		Type:      StateUpdate,
		Payload:   map[string]interface{}{"action": "REPORT_CAPABILITIES"},
	})

	return ch
}

func (b *A2ABroker) UnregisterAgent(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, ok := b.agents[id]; ok {
		close(ch)
		delete(b.agents, id)
		delete(b.heartbeats, id)
	}
}

func (b *A2ABroker) SetEventBus(bus interface {
	EmitEvent(eventType string, source string, payload interface{})
}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.bus = bus
}

func (b *A2ABroker) RouteMessage(msg A2AMessage) {
	if b.logger != nil {
		b.logger.LogMessage(msg)
	}

	b.mu.RLock()
	bus := b.bus
	proc := b.signalProcessor
	b.mu.RUnlock()

	if proc != nil {
		proc.ProcessSignal(context.Background(), msg)
	}

	if bus != nil {
		bus.EmitEvent("a2a:signal", "A2ABroker", map[string]interface{}{
			"message": msg,
		})
	}

	b.mu.Lock()
	if msg.Sender != "BROKER" {
		b.heartbeats[msg.Sender] = time.Now().UnixMilli()
	}

	if msg.Type == Heartbeat {
		b.mu.Unlock()
		return
	}

	if msg.ReplyTo != "" {
		if ch, ok := b.pendingResponses[msg.ReplyTo]; ok {
			select {
			case ch <- msg:
			default:
			}
		}
	}

	b.history = append(b.history, msg)
	if len(b.history) > 1000 {
		b.history = b.history[1:]
	}
	b.mu.Unlock()

	b.mu.RLock()
	defer b.mu.RUnlock()

	if msg.Recipient != "" {
		if ch, ok := b.agents[msg.Recipient]; ok {
			select {
			case ch <- msg:
			default:
			}
		}
	} else {
		for id, ch := range b.agents {
			if id == msg.Sender { continue }
			select {
			case ch <- msg:
			default:
			}
		}
	}
}

func (b *A2ABroker) GetHistory() []A2AMessage {
	b.mu.RLock()
	defer b.mu.RUnlock()
	h := make([]A2AMessage, len(b.history))
	copy(h, b.history)
	return h
}

func (b *A2ABroker) ListAgents() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	agents := make([]string, 0, len(b.agents))
	for id := range b.agents {
		agents = append(agents, id)
	}
	return agents
}

func _nowMillis_deprecated() int64 {
	return time.Now().UnixMilli()
}
