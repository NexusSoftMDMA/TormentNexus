package interop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/borghq/borg-go/internal/lockfile"
)

var defaultTRPCBases = []string{
	"http://127.0.0.1:4100/trpc",
	"http://127.0.0.1:4001/trpc",
	"http://127.0.0.1:3001/trpc",
}

type UpstreamCallResult struct {
	BaseURL string          `json:"baseUrl"`
	Data    json.RawMessage `json:"data"`
}

func ResolveTRPCBases(mainLockPath string) []string {
	configured := strings.TrimSpace(os.Getenv("BORG_TRPC_UPSTREAM"))
	bases := make([]string, 0, len(defaultTRPCBases)+2)
	if lockedBase := resolveLockedTRPCBase(mainLockPath); lockedBase != "" {
		bases = append(bases, lockedBase)
	}
	if configured != "" {
		bases = append(bases, configured)
	}
	bases = append(bases, defaultTRPCBases...)

	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(bases))
	for _, base := range bases {
		base = strings.TrimSpace(strings.TrimRight(base, "/"))
		if base == "" {
			continue
		}
		if _, ok := seen[base]; ok {
			continue
		}
		seen[base] = struct{}{}
		normalized = append(normalized, base)
	}
	return normalized
}

func CallTRPCProcedure(ctx context.Context, mainLockPath string, procedure string, payload any) (UpstreamCallResult, error) {
	requestBody, err := json.Marshal(map[string]any{"json": payload})
	if err != nil {
		return UpstreamCallResult{}, err
	}

	var lastErr error
	client := &http.Client{Timeout: 10 * time.Second}
	for _, base := range ResolveTRPCBases(mainLockPath) {
		targetURL := strings.TrimRight(base, "/") + "/" + strings.TrimLeft(procedure, "/")
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(requestBody))
		if err != nil {
			lastErr = err
			continue
		}
		request.Header.Set("content-type", "application/json")

		response, err := client.Do(request)
		if err != nil {
			lastErr = err
			continue
		}

		body, readErr := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if response.StatusCode >= http.StatusBadRequest {
			lastErr = fmt.Errorf("upstream %s returned %d: %s", targetURL, response.StatusCode, strings.TrimSpace(string(body)))
			continue
		}

		data, extractErr := extractTRPCData(body)
		if extractErr != nil {
			lastErr = extractErr
			continue
		}
		return UpstreamCallResult{BaseURL: base, Data: data}, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no TypeScript control-plane upstreams available")
	}
	return UpstreamCallResult{}, lastErr
}

func resolveLockedTRPCBase(mainLockPath string) string {
	record, err := lockfile.Read(mainLockPath)
	if err != nil || record.Port <= 0 {
		return ""
	}

	host := strings.TrimSpace(record.Host)
	switch host {
	case "", "0.0.0.0", "::", "[::]":
		host = "127.0.0.1"
	}

	return fmt.Sprintf("http://%s:%d/trpc", host, record.Port)
}

func extractTRPCData(body []byte) (json.RawMessage, error) {
	var single struct {
		Result *struct {
			Data json.RawMessage `json:"data"`
		} `json:"result"`
		Error any `json:"error"`
	}
	if err := json.Unmarshal(body, &single); err == nil && single.Result != nil {
		return unwrapTRPCData(single.Result.Data), nil
	}

	var batched []struct {
		Result *struct {
			Data json.RawMessage `json:"data"`
		} `json:"result"`
		Error any `json:"error"`
	}
	if err := json.Unmarshal(body, &batched); err == nil && len(batched) > 0 && batched[0].Result != nil {
		return unwrapTRPCData(batched[0].Result.Data), nil
	}

	return nil, fmt.Errorf("unexpected tRPC response shape")
}

func unwrapTRPCData(data json.RawMessage) json.RawMessage {
	var wrapped struct {
		JSON json.RawMessage `json:"json"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && len(wrapped.JSON) > 0 {
		return wrapped.JSON
	}
	return data
}
