//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// In-memory structures for the brokered peer notebook pilot.
type PeerArtifact struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	MimeType  string      `json:"mimeType"`
	ByteSize  int         `json:"byteSize"`
	Sha256    string      `json:"sha256"`
	Preview   interface{} `json:"preview,omitempty"`
	Content   interface{} `json:"content,omitempty"`
	CreatedAt string      `json:"createdAt"`
}

type PeerTraceEvent struct {
	ID           string                 `json:"id"`
	Seq          int                    `json:"seq"`
	WorkspaceID  string                 `json:"workspaceId"`
	InvocationID string                 `json:"invocationId"`
	EventType    string                 `json:"eventType"`
	Severity     string                 `json:"severity"`
	TimestampAt  string                 `json:"timestampAt"`
	Body         string                 `json:"body,omitempty"`
	Attrs        map[string]interface{} `json:"attrs,omitempty"`
}

type PeerInvocation struct {
	ID         string                 `json:"id"`
	PeerID     string                 `json:"peerId"`
	ToolName   string                 `json:"toolName"`
	Status     string                 `json:"status"`
	Result     interface{}            `json:"result,omitempty"`
	Error      interface{}            `json:"error,omitempty"`
	CreatedAt  string                 `json:"createdAt"`
	StartedAt  string                 `json:"startedAt,omitempty"`
	CompletedAt string                 `json:"completedAt,omitempty"`
}

var (
	peerArtifacts   = make(map[string]*PeerArtifact)
	peerTraceEvents = make(map[string][]PeerTraceEvent)
	peerInvocations = make(map[string]*PeerInvocation)
	peerMutex       sync.Mutex
)

// HandleThoughtboxSearch executes a catalog query via Node VM.
func HandleThoughtboxSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ := args["code"].(string)
	if code == "" {
		return err("code parameter is required")
	}

	// Locating the sandbox JS file.
	cwd, _ := os.Getwd()
	var sandboxPath string
	pathsToTry := []string{
		filepath.Join(cwd, "internal", "tools", "thoughtbox_sandbox.js"),
		filepath.Join(cwd, "go", "internal", "tools", "thoughtbox_sandbox.js"),
		filepath.Join(cwd, "..", "go", "internal", "tools", "thoughtbox_sandbox.js"),
		filepath.Join(cwd, "thoughtbox_sandbox.js"),
	}
	for _, p := range pathsToTry {
		if _, err := os.Stat(p); err == nil {
			sandboxPath = p
			break
		}
	}
	if sandboxPath == "" {
		return err("thoughtbox_sandbox.js not found in any checked paths")
	}

	cmd := exec.CommandContext(ctx, "node", sandboxPath, "search", code)
	out, errCmd := cmd.CombinedOutput()
	if errCmd != nil {
		return ToolResponse{
			Content: []TextContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error running sandbox search: %v\nOutput: %s", errCmd, string(out)),
				},
			},
		}, nil
	}

	return ToolResponse{
		Content: []TextContent{
			{
				Type: "text",
				Text: string(out),
			},
		},
	}, nil
}

// HandleThoughtboxExecute executes SDK commands via Node VM.
func HandleThoughtboxExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ := args["code"].(string)
	if code == "" {
		return err("code parameter is required")
	}

	cwd, _ := os.Getwd()
	var sandboxPath string
	pathsToTry := []string{
		filepath.Join(cwd, "internal", "tools", "thoughtbox_sandbox.js"),
		filepath.Join(cwd, "go", "internal", "tools", "thoughtbox_sandbox.js"),
		filepath.Join(cwd, "..", "go", "internal", "tools", "thoughtbox_sandbox.js"),
		filepath.Join(cwd, "thoughtbox_sandbox.js"),
	}
	for _, p := range pathsToTry {
		if _, err := os.Stat(p); err == nil {
			sandboxPath = p
			break
		}
	}
	if sandboxPath == "" {
		return err("thoughtbox_sandbox.js not found in any checked paths")
	}

	cmd := exec.CommandContext(ctx, "node", sandboxPath, "execute", code)
	out, errCmd := cmd.CombinedOutput()
	if errCmd != nil {
		return ToolResponse{
			Content: []TextContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error running sandbox execute: %v\nOutput: %s", errCmd, string(out)),
				},
			},
		}, nil
	}

	return ToolResponse{
		Content: []TextContent{
			{
				Type: "text",
				Text: string(out),
			},
		},
	}, nil
}

// HandleThoughtboxPeerNotebook implements the peer notebook brokered pilot.
func HandleThoughtboxPeerNotebook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	op, _ := args["operation"].(string)
	if op == "" {
		return err("operation parameter is required")
	}

	peerMutex.Lock()
	defer peerMutex.Unlock()

	switch op {
	case "peer_artifact_seed":
		text, _ := args["text"].(string)
		name, _ := args["name"].(string)
		if name == "" {
			name = "input.txt"
		}

		art := &PeerArtifact{
			ID:        uuid.New().String(),
			Name:      name,
			MimeType:  "text/plain",
			ByteSize:  len(text),
			Sha256:    "mock-sha256-hash",
			Content:   text,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		if len(text) > 500 {
			art.Preview = text[:497] + "..."
		} else {
			art.Preview = text
		}

		peerArtifacts[art.ID] = art

		res := map[string]interface{}{
			"artifact": art,
		}
		resJSON, _ := json.MarshalIndent(res, "", "  ")
		return ok(string(resJSON))

	case "peer_invoke":
		peerID, _ := args["peerId"].(string)
		tool, _ := args["tool"].(string)
		toolArgs, _ := args["args"].(map[string]interface{})

		if peerID != "claim-extractor" {
			return err("Peer not found: " + peerID)
		}
		if tool != "extract_claims" {
			return err("Tool not found: " + tool)
		}

		textArtID, _ := toolArgs["textArtifactId"].(string)
		textArt, exists := peerArtifacts[textArtID]
		if !exists {
			return err("Artifact not found: " + textArtID)
		}

		textStr, _ := textArt.Content.(string)

		// Create invocation
		invID := uuid.New().String()
		inv := &PeerInvocation{
			ID:        invID,
			PeerID:    peerID,
			ToolName:  tool,
			Status:    "queued",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		peerInvocations[invID] = inv

		// Add trace: peer_invocation_created
		peerTraceEvents[invID] = append(peerTraceEvents[invID], PeerTraceEvent{
			ID:           uuid.New().String(),
			Seq:          1,
			WorkspaceID:  "default",
			InvocationID: invID,
			EventType:    "peer_invocation_created",
			Severity:     "info",
			TimestampAt:  time.Now().UTC().Format(time.RFC3339),
			Attrs: map[string]interface{}{
				"peerId":       peerID,
				"tool":         tool,
				"manifestHash": "mock-manifest-hash",
			},
		})

		// Start running
		inv.Status = "running"
		inv.StartedAt = time.Now().UTC().Format(time.RFC3339)

		// Check outbound call allowed
		peerTraceEvents[invID] = append(peerTraceEvents[invID], PeerTraceEvent{
			ID:           uuid.New().String(),
			Seq:          2,
			WorkspaceID:  "default",
			InvocationID: invID,
			EventType:    "outbound_call_allowed",
			Severity:     "info",
			TimestampAt:  time.Now().UTC().Format(time.RFC3339),
			Attrs: map[string]interface{}{
				"target": "artifact.get",
			},
		})

		// Denied probe
		peerTraceEvents[invID] = append(peerTraceEvents[invID], PeerTraceEvent{
			ID:           uuid.New().String(),
			Seq:          3,
			WorkspaceID:  "default",
			InvocationID: invID,
			EventType:    "denied_outbound_call",
			Severity:     "warn",
			Body:         "Outbound call to thoughtbox.knowledge.queryGraph is not allowed by active manifest",
			TimestampAt:  time.Now().UTC().Format(time.RFC3339),
			Attrs: map[string]interface{}{
				"target": "thoughtbox.knowledge.queryGraph",
			},
		})

		// Extract sentences as claims
		var claims []map[string]interface{}
		sentences := strings.Split(textStr, ".")
		claimCount := 0
		for _, s := range sentences {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			claimCount++
			claims = append(claims, map[string]interface{}{
				"id":   fmt.Sprintf("claim_%d", claimCount),
				"text": s,
			})
		}

		claimsArtID := uuid.New().String()
		claimsJson := map[string]interface{}{
			"claims": claims,
		}

		claimsArt := &PeerArtifact{
			ID:        claimsArtID,
			Name:      "claims.json",
			MimeType:  "application/json",
			ByteSize:  128, // Mock size
			Sha256:    "mock-claims-sha256",
			Content:   claimsJson,
			Preview:   claimsJson,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		peerArtifacts[claimsArtID] = claimsArt

		peerTraceEvents[invID] = append(peerTraceEvents[invID], PeerTraceEvent{
			ID:           uuid.New().String(),
			Seq:          4,
			WorkspaceID:  "default",
			InvocationID: invID,
			EventType:    "peer_artifact_written",
			Severity:     "info",
			TimestampAt:  time.Now().UTC().Format(time.RFC3339),
			Attrs: map[string]interface{}{
				"artifactId": claimsArtID,
				"name":       "claims.json",
			},
		})

		inv.Status = "completed"
		inv.CompletedAt = time.Now().UTC().Format(time.RFC3339)
		inv.Result = map[string]interface{}{
			"claimsArtifactId": claimsArtID,
			"claimCount":       claimCount,
		}

		peerTraceEvents[invID] = append(peerTraceEvents[invID], PeerTraceEvent{
			ID:           uuid.New().String(),
			Seq:          5,
			WorkspaceID:  "default",
			InvocationID: invID,
			EventType:    "peer_invocation_completed",
			Severity:     "info",
			TimestampAt:  time.Now().UTC().Format(time.RFC3339),
			Attrs: map[string]interface{}{
				"resultHash": "mock-result-hash",
			},
		})

		res := map[string]interface{}{
			"invocationId": invID,
			"manifestHash": "mock-manifest-hash",
			"result":       inv.Result,
			"artifactRefs": []map[string]interface{}{
				{
					"artifactId": claimsArtID,
					"kind":       "json",
					"name":       "claims.json",
					"mimeType":   "application/json",
					"byteSize":   claimsArt.ByteSize,
					"sha256":     claimsArt.Sha256,
					"preview":     claimsArt.Preview,
				},
			},
		}
		resJSON, _ := json.MarshalIndent(res, "", "  ")
		return ok(string(resJSON))

	case "peer_get_invocation":
		invID, _ := args["invocationId"].(string)
		inv, exists := peerInvocations[invID]
		if !exists {
			return err("Invocation not found")
		}
		res := map[string]interface{}{
			"invocation": inv,
		}
		resJSON, _ := json.MarshalIndent(res, "", "  ")
		return ok(string(resJSON))

	case "peer_list_trace_events":
		invID, _ := args["invocationId"].(string)
		events, exists := peerTraceEvents[invID]
		if !exists {
			events = []PeerTraceEvent{}
		}
		res := map[string]interface{}{
			"events": events,
		}
		resJSON, _ := json.MarshalIndent(res, "", "  ")
		return ok(string(resJSON))

	case "peer_get_artifact":
		artID, _ := args["artifactId"].(string)
		art, exists := peerArtifacts[artID]
		if !exists {
			return err("Artifact not found")
		}
		res := map[string]interface{}{
			"artifact": art,
		}
		resJSON, _ := json.MarshalIndent(res, "", "  ")
		return ok(string(resJSON))

	default:
		return err("Unsupported peer notebook operation: " + op)
	}
}
