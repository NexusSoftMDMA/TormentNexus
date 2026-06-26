package buffer

import (
	"sync"
)

type BufferedMessage struct {
	ID        string      `json:"id"`
	Timestamp int64       `json:"timestamp"`
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
}

type ResilientBuffer struct {
	mu         sync.RWMutex
	messages   []BufferedMessage
	maxSize    int
	lastIndex  int64
}

func NewResilientBuffer(maxSize int) *ResilientBuffer {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &ResilientBuffer{
		messages: make([]BufferedMessage, 0, maxSize),
		maxSize:  maxSize,
	}
}

func (b *ResilientBuffer) Push(msg BufferedMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.messages) >= b.maxSize {
		b.messages = b.messages[1:]
	}
	b.messages = append(b.messages, msg)
}

func (b *ResilientBuffer) GetSince(timestamp int64) []BufferedMessage {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var result []BufferedMessage
	for _, m := range b.messages {
		if m.Timestamp > timestamp {
			result = append(result, m)
		}
	}
	return result
}

func (b *ResilientBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = nil
}
