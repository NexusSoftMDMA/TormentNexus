// Package toon implements the TOON (Thread-Oriented Object Notation) format,
// a compact binary serialization for conversation context snapshots.
//
// TOON is designed for efficient context compression and cross-session
// portability in the TormentNexus ecosystem. It encodes structured thread data
// (messages, tool calls, memory references, and metadata) into a compact
// binary representation that can be persisted, transmitted, and decoded
// with minimal overhead.
//
// Format layout:
//
//	[Header: 8 bytes magic + 4 bytes version + 4 bytes flags]
//	[Thread Metadata Section: length-prefixed JSON]
//	[Message Entries: count + sequential entries]
//	[Memory References Section: count + sequential refs]
//	[Footer: 4 bytes CRC32 checksum]
//
// Each message entry uses a type-tagged compact encoding:
//
//	[1 byte type] [8 bytes timestamp] [4 bytes content length] [N bytes content]
//	[2 bytes tool_calls_count] [tool_call entries...]
//	[2 bytes metadata_kv_count] [metadata entries...]
package toon

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"time"
)

// Magic bytes identifying a TOON stream: "TOON" + format revision (0x01, 0x00)
var Magic = [8]byte{'T', 'O', 'O', 'N', 0x01, 0x00, 0x00, 0x00}

// Version is the current TOON format version.
const Version uint32 = 1

// Flags for the TOON header.
const (
	FlagEncrypted  uint32 = 1 << 0 // Payload is encrypted (AES-GCM)
	FlagCompressed uint32 = 1 << 1 // Payload is compressed (zstd)
	FlagSigned     uint32 = 1 << 2 // Payload has a signature block
)

// Role identifies the speaker in a thread message.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

// MessageType classifies the content of a thread message.
type MessageType uint8

const (
	TypeText      MessageType = 0x01
	TypeToolCall  MessageType = 0x02
	TypeToolResult MessageType = 0x03
	TypeImage     MessageType = 0x04
	TypeCode      MessageType = 0x05
	TypeMemory    MessageType = 0x06
	TypeThinking  MessageType = 0x07
	TypeSummary   MessageType = 0x08
)

// ToolCall represents a single tool invocation within a message.
type ToolCall struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Args     map[string]interface{} `json:"args,omitempty"`
	Result   interface{}            `json:"result,omitempty"`
	Duration int64                  `json:"duration_ms,omitempty"`
	Error    string                 `json:"error,omitempty"`
}

// Message is a single entry in a TOON thread.
type Message struct {
	Role      Role        `json:"role"`
	Type      MessageType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Content   string      `json:"content"`
	ToolCalls []ToolCall  `json:"tool_calls,omitempty"`
	Metadata  MapEntries  `json:"metadata,omitempty"`
}

// MapEntries is an ordered collection of key-value pairs,
// preserving insertion order unlike a plain map.
type MapEntries []MapEntry

// MapEntry is a single key-value pair.
type MapEntry struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// MemoryRef is a reference to an L2 Vault memory record.
type MemoryRef struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Content    string  `json:"content"`
	Importance float64 `json:"importance"`
	HeatScore  float64 `json:"heat_score"`
}

// Thread is the top-level TOON container.
type Thread struct {
	ID        string       `json:"id"`
	SessionID string       `json:"session_id"`
	Agent     string       `json:"agent,omitempty"`
	Model     string       `json:"model,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	Tags      []string     `json:"tags,omitempty"`
	Messages  []Message    `json:"messages"`
	Memories  []MemoryRef  `json:"memories,omitempty"`
	Metadata  MapEntries   `json:"metadata,omitempty"`
}

// Header is the binary TOON file header.
type Header struct {
	Magic   [8]byte
	Version uint32
	Flags   uint32
}

// Encoder writes TOON format to a stream.
type Encoder struct {
	w io.Writer
}

// NewEncoder creates a TOON encoder writing to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode serializes a Thread into TOON binary format.
func (e *Encoder) Encode(t *Thread) error {
	// Build the payload in a buffer first so we can compute CRC
	var buf bytes.Buffer

	// Write header
	hdr := Header{Magic: Magic, Version: Version, Flags: 0}
	if err := binary.Write(&buf, binary.LittleEndian, &hdr); err != nil {
		return fmt.Errorf("toon encode header: %w", err)
	}

	// Write thread metadata as length-prefixed JSON
	meta := map[string]interface{}{
		"id":         t.ID,
		"session_id": t.SessionID,
		"agent":      t.Agent,
		"model":      t.Model,
		"created_at": t.CreatedAt.UnixMilli(),
		"tags":       t.Tags,
	}
	for _, entry := range t.Metadata {
		meta[entry.Key] = entry.Value
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("toon encode metadata: %w", err)
	}
	if err := binary.Write(&buf, binary.LittleEndian, uint32(len(metaJSON))); err != nil {
		return fmt.Errorf("toon encode metadata length: %w", err)
	}
	if _, err := buf.Write(metaJSON); err != nil {
		return fmt.Errorf("toon encode metadata body: %w", err)
	}

	// Write message count
	if err := binary.Write(&buf, binary.LittleEndian, uint32(len(t.Messages))); err != nil {
		return fmt.Errorf("toon encode message count: %w", err)
	}

	// Write each message
	for _, msg := range t.Messages {
		if err := encodeMessage(&buf, msg); err != nil {
			return fmt.Errorf("toon encode message: %w", err)
		}
	}

	// Write memory references count
	if err := binary.Write(&buf, binary.LittleEndian, uint32(len(t.Memories))); err != nil {
		return fmt.Errorf("toon encode memory count: %w", err)
	}

	// Write each memory reference
	for _, mem := range t.Memories {
		if err := encodeMemoryRef(&buf, mem); err != nil {
			return fmt.Errorf("toon encode memory ref: %w", err)
		}
	}

	// Compute CRC32 of everything so far
	checksum := crc32.ChecksumIEEE(buf.Bytes())
	if err := binary.Write(&buf, binary.LittleEndian, checksum); err != nil {
		return fmt.Errorf("toon encode checksum: %w", err)
	}

	// Write the complete buffer
	_, err = e.w.Write(buf.Bytes())
	return err
}

func encodeMessage(w *bytes.Buffer, msg Message) error {
	// Type byte
	if err := w.WriteByte(byte(msg.Type)); err != nil {
		return err
	}

	// Role as length-prefixed string
	if err := encodeString(w, string(msg.Role)); err != nil {
		return err
	}

	// Timestamp as milliseconds since epoch
	if err := binary.Write(w, binary.LittleEndian, msg.Timestamp.UnixMilli()); err != nil {
		return err
	}

	// Content as length-prefixed string
	if err := encodeString(w, msg.Content); err != nil {
		return err
	}

	// Tool calls count
	if err := binary.Write(w, binary.LittleEndian, uint16(len(msg.ToolCalls))); err != nil {
		return err
	}

	// Each tool call as length-prefixed JSON
	for _, tc := range msg.ToolCalls {
		tcJSON, err := json.Marshal(tc)
		if err != nil {
			return fmt.Errorf("toon encode tool call: %w", err)
		}
		if err := encodeBytes(w, tcJSON); err != nil {
			return err
		}
	}

	// Metadata count
	if err := binary.Write(w, binary.LittleEndian, uint16(len(msg.Metadata))); err != nil {
		return err
	}

	// Each metadata entry
	for _, entry := range msg.Metadata {
		if err := encodeString(w, entry.Key); err != nil {
			return err
		}
		valJSON, err := json.Marshal(entry.Value)
		if err != nil {
			return fmt.Errorf("toon encode metadata value: %w", err)
		}
		if err := encodeBytes(w, valJSON); err != nil {
			return err
		}
	}

	return nil
}

func encodeMemoryRef(w *bytes.Buffer, mem MemoryRef) error {
	if err := encodeString(w, mem.ID); err != nil {
		return err
	}
	if err := encodeString(w, mem.Type); err != nil {
		return err
	}
	if err := encodeString(w, mem.Content); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, mem.Importance); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, mem.HeatScore); err != nil {
		return err
	}
	return nil
}

// Decoder reads TOON format from a stream.
type Decoder struct {
	r io.Reader
}

// NewDecoder creates a TOON decoder reading from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode deserializes a Thread from TOON binary format.
func (d *Decoder) Decode() (*Thread, error) {
	// Read entire stream for CRC validation
	data, err := io.ReadAll(d.r)
	if err != nil {
		return nil, fmt.Errorf("toon decode read: %w", err)
	}

	if len(data) < 20 { // 16 header + 4 checksum minimum
		return nil, errors.New("toon decode: data too short")
	}

	// Validate CRC: everything except the last 4 bytes
	payload := data[:len(data)-4]
	storedChecksum := binary.LittleEndian.Uint32(data[len(data)-4:])
	computedChecksum := crc32.ChecksumIEEE(payload)
	if storedChecksum != computedChecksum {
		return nil, fmt.Errorf("toon decode: checksum mismatch (stored=%08x, computed=%08x)", storedChecksum, computedChecksum)
	}

	buf := bytes.NewReader(payload)

	// Read header
	var hdr Header
	if err := binary.Read(buf, binary.LittleEndian, &hdr); err != nil {
		return nil, fmt.Errorf("toon decode header: %w", err)
	}

	// Validate magic
	if hdr.Magic != Magic {
		return nil, fmt.Errorf("toon decode: invalid magic %x (expected %x)", hdr.Magic, Magic)
	}

	// Validate version
	if hdr.Version > Version {
		return nil, fmt.Errorf("toon decode: unsupported version %d (max %d)", hdr.Version, Version)
	}

	// Read thread metadata
	var metaLen uint32
	if err := binary.Read(buf, binary.LittleEndian, &metaLen); err != nil {
		return nil, fmt.Errorf("toon decode metadata length: %w", err)
	}
	metaJSON := make([]byte, metaLen)
	if _, err := io.ReadFull(buf, metaJSON); err != nil {
		return nil, fmt.Errorf("toon decode metadata body: %w", err)
	}

	var meta map[string]interface{}
	if err := json.Unmarshal(metaJSON, &meta); err != nil {
		return nil, fmt.Errorf("toon decode metadata parse: %w", err)
	}

	thread := &Thread{}

	// Extract known fields from metadata
	if v, ok := meta["id"].(string); ok {
		thread.ID = v
	}
	if v, ok := meta["session_id"].(string); ok {
		thread.SessionID = v
	}
	if v, ok := meta["agent"].(string); ok {
		thread.Agent = v
	}
	if v, ok := meta["model"].(string); ok {
		thread.Model = v
	}
	if v, ok := meta["created_at"].(float64); ok {
		thread.CreatedAt = time.UnixMilli(int64(v))
	}
	if v, ok := meta["tags"].([]interface{}); ok {
		for _, tag := range v {
			if s, ok := tag.(string); ok {
				thread.Tags = append(thread.Tags, s)
			}
		}
	}

	// Store remaining metadata as MapEntries
	knownKeys := map[string]bool{"id": true, "session_id": true, "agent": true, "model": true, "created_at": true, "tags": true}
	for k, v := range meta {
		if !knownKeys[k] {
			thread.Metadata = append(thread.Metadata, MapEntry{Key: k, Value: v})
		}
	}

	// Read messages
	var msgCount uint32
	if err := binary.Read(buf, binary.LittleEndian, &msgCount); err != nil {
		return nil, fmt.Errorf("toon decode message count: %w", err)
	}

	thread.Messages = make([]Message, 0, msgCount)
	for i := uint32(0); i < msgCount; i++ {
		msg, err := decodeMessage(buf)
		if err != nil {
			return nil, fmt.Errorf("toon decode message %d: %w", i, err)
		}
		thread.Messages = append(thread.Messages, msg)
	}

	// Read memory references
	var memCount uint32
	if err := binary.Read(buf, binary.LittleEndian, &memCount); err != nil {
		return nil, fmt.Errorf("toon decode memory count: %w", err)
	}

	thread.Memories = make([]MemoryRef, 0, memCount)
	for i := uint32(0); i < memCount; i++ {
		mem, err := decodeMemoryRef(buf)
		if err != nil {
			return nil, fmt.Errorf("toon decode memory ref %d: %w", i, err)
		}
		thread.Memories = append(thread.Memories, mem)
	}

	return thread, nil
}

func decodeMessage(buf *bytes.Reader) (Message, error) {
	var msg Message

	// Type byte
	typeByte, err := buf.ReadByte()
	if err != nil {
		return msg, err
	}
	msg.Type = MessageType(typeByte)

	// Role
	role, err := decodeString(buf)
	if err != nil {
		return msg, err
	}
	msg.Role = Role(role)

	// Timestamp
	var tsMilli int64
	if err := binary.Read(buf, binary.LittleEndian, &tsMilli); err != nil {
		return msg, err
	}
	msg.Timestamp = time.UnixMilli(tsMilli)

	// Content
	msg.Content, err = decodeString(buf)
	if err != nil {
		return msg, err
	}

	// Tool calls
	var tcCount uint16
	if err := binary.Read(buf, binary.LittleEndian, &tcCount); err != nil {
		return msg, err
	}

	msg.ToolCalls = make([]ToolCall, 0, tcCount)
	for i := uint16(0); i < tcCount; i++ {
		tcJSON, err := decodeBytes(buf)
		if err != nil {
			return msg, fmt.Errorf("tool call %d: %w", i, err)
		}
		var tc ToolCall
		if err := json.Unmarshal(tcJSON, &tc); err != nil {
			return msg, fmt.Errorf("tool call %d parse: %w", i, err)
		}
		msg.ToolCalls = append(msg.ToolCalls, tc)
	}

	// Metadata
	var mdCount uint16
	if err := binary.Read(buf, binary.LittleEndian, &mdCount); err != nil {
		return msg, err
	}

	msg.Metadata = make(MapEntries, 0, mdCount)
	for i := uint16(0); i < mdCount; i++ {
		key, err := decodeString(buf)
		if err != nil {
			return msg, fmt.Errorf("metadata key %d: %w", i, err)
		}
		valJSON, err := decodeBytes(buf)
		if err != nil {
			return msg, fmt.Errorf("metadata value %d: %w", i, err)
		}
		var val interface{}
		if err := json.Unmarshal(valJSON, &val); err != nil {
			return msg, fmt.Errorf("metadata value %d parse: %w", i, err)
		}
		msg.Metadata = append(msg.Metadata, MapEntry{Key: key, Value: val})
	}

	return msg, nil
}

func decodeMemoryRef(buf *bytes.Reader) (MemoryRef, error) {
	var mem MemoryRef
	var err error

	mem.ID, err = decodeString(buf)
	if err != nil {
		return mem, err
	}
	mem.Type, err = decodeString(buf)
	if err != nil {
		return mem, err
	}
	mem.Content, err = decodeString(buf)
	if err != nil {
		return mem, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &mem.Importance); err != nil {
		return mem, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &mem.HeatScore); err != nil {
		return mem, err
	}

	return mem, nil
}

// EncodeToBytes is a convenience function that encodes a Thread to a byte slice.
func EncodeToBytes(t *Thread) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(t); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeFromBytes is a convenience function that decodes a Thread from a byte slice.
func DecodeFromBytes(data []byte) (*Thread, error) {
	dec := NewDecoder(bytes.NewReader(data))
	return dec.Decode()
}

// Encrypt encrypts a TOON payload with AES-GCM using the provided key.
// The key must be 32 bytes (AES-256).
func Encrypt(payload []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("toon encrypt: key must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("toon encrypt: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("toon encrypt gcm: %w", err)
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("toon encrypt nonce: %w", err)
	}

	// Prepend nonce to ciphertext
	ciphertext := aesgcm.Seal(nil, nonce, payload, nil)
	return append(nonce, ciphertext...), nil
}

// Decrypt decrypts a TOON payload that was encrypted with Encrypt.
func Decrypt(data []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("toon decrypt: key must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("toon decrypt: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("toon decrypt gcm: %w", err)
	}

	nonceSize := aesgcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("toon decrypt: data too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("toon decrypt open: %w", err)
	}

	return plaintext, nil
}

// PrunableThread creates a compressed snapshot of a thread,
// keeping only the last N messages plus all tool results and memory references.
// This is the core of TOON context compression.
func PrunableThread(t *Thread, keepLast int) *Thread {
	if keepLast >= len(t.Messages) {
		return t
	}

	result := &Thread{
		ID:        t.ID,
		SessionID: t.SessionID,
		Agent:     t.Agent,
		Model:     t.Model,
		CreatedAt: t.CreatedAt,
		Tags:      t.Tags,
		Memories:  t.Memories,
		Metadata:  t.Metadata,
	}

	// Always keep system messages and tool results
	var preserved []Message
	var tail []Message

	for _, msg := range t.Messages {
		if msg.Role == RoleSystem || msg.Type == TypeToolResult {
			preserved = append(preserved, msg)
		}
	}

	// Keep the last N messages
	start := len(t.Messages) - keepLast
	if start < 0 {
		start = 0
	}
	tail = t.Messages[start:]

	// Merge, deduping by timestamp+role
	seen := make(map[string]bool)
	var merged []Message

	for _, msg := range preserved {
		key := fmt.Sprintf("%d:%s:%s", msg.Timestamp.UnixMilli(), msg.Role, truncate(msg.Content, 64))
		if !seen[key] {
			seen[key] = true
			merged = append(merged, msg)
		}
	}

	for _, msg := range tail {
		key := fmt.Sprintf("%d:%s:%s", msg.Timestamp.UnixMilli(), msg.Role, truncate(msg.Content, 64))
		if !seen[key] {
			seen[key] = true
			merged = append(merged, msg)
		}
	}

	result.Messages = merged
	return result
}

// SummarizeThread creates a summary TOON from a thread,
// replacing all messages with a single TypeSummary message
// while preserving all memory references.
func SummarizeThread(t *Thread, summary string) *Thread {
	return &Thread{
		ID:        t.ID,
		SessionID: t.SessionID,
		Agent:     t.Agent,
		Model:     t.Model,
		CreatedAt: t.CreatedAt,
		Tags:      t.Tags,
		Memories:  t.Memories,
		Metadata: append(t.Metadata, MapEntry{
			Key:   "original_message_count",
			Value: len(t.Messages),
		}),
		Messages: []Message{
			{
				Role:      RoleSystem,
				Type:      TypeSummary,
				Timestamp: time.Now(),
				Content:   summary,
			},
		},
	}
}

// Helper functions

func encodeString(w *bytes.Buffer, s string) error {
	if err := binary.Write(w, binary.LittleEndian, uint32(len(s))); err != nil {
		return err
	}
	_, err := w.WriteString(s)
	return err
}

func decodeString(buf *bytes.Reader) (string, error) {
	var length uint32
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	s := make([]byte, length)
	if _, err := io.ReadFull(buf, s); err != nil {
		return "", err
	}
	return string(s), nil
}

func encodeBytes(w *bytes.Buffer, b []byte) error {
	if err := binary.Write(w, binary.LittleEndian, uint32(len(b))); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}

func decodeBytes(buf *bytes.Reader) ([]byte, error) {
	var length uint32
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	b := make([]byte, length)
	if _, err := io.ReadFull(buf, b); err != nil {
		return nil, err
	}
	return b, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
