package interop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/borghq/borg-go/internal/config"
	"github.com/borghq/borg-go/internal/lockfile"
)

// sharedTRPCClient returns a singleton HTTP client with connection pooling
// tuned for concurrent upstream calls to the TypeScript control plane.
var sharedClientOnce sync.Once
var sharedClientInst *http.Client

func sharedTRPCClient() *http.Client {
	sharedClientOnce.Do(func() {
		transport := &http.Transport{
			MaxIdleConns:        20,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     10,
			IdleConnTimeout:     30 * time.Second,
		}
		sharedClientInst = &http.Client{
			Timeout:   1500 * time.Millisecond,
			Transport: transport,
		}
	})
	return sharedClientInst
}

var defaultTRPCBases = []string{
	"http://127.0.0.1:4100/trpc",
	"http://127.0.0.1:4000/trpc",
	"http://127.0.0.1:3847/trpc",
}

func DefaultTRPCBasesFromDiscovery() []string {
	sd := config.DefaultServiceDiscovery()
	if len(sd.TRPCUpstreamURLs) > 0 {
		return sd.TRPCUpstreamURLs
	}
	return defaultTRPCBases
}

type UpstreamCallResult struct {
	BaseURL string          `json:"baseUrl"`
	Data    json.RawMessage `json:"data"`
}

// ResolveTRPCBases returns the ordered list of tRPC upstream base URLs.
// If BORG_TRPC_UPSTREAM is set, it is used exclusively (no lockfile or defaults).
// Otherwise, lockfile-recorded base is tried first, then discovery defaults.
func ResolveTRPCBases(mainLockPath string) []string {
	configured := strings.TrimSpace(os.Getenv("BORG_TRPC_UPSTREAM"))
	if configured != "" {
		// When explicit upstream is set, use it exclusively —
		// this prevents tests from accidentally hitting real servers.
		return []string{configured}
	}

	bases := make([]string, 0, len(DefaultTRPCBasesFromDiscovery())+2)
	if lockedBase := resolveLockedTRPCBase(mainLockPath); lockedBase != "" {
		bases = append(bases, lockedBase)
	}
	bases = append(bases, DefaultTRPCBasesFromDiscovery()...)

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
	client := sharedTRPCClient()

	bases := ResolveTRPCBases(mainLockPath)
	if len(bases) == 0 {
		return UpstreamCallResult{}, fmt.Errorf("no TypeScript control-plane upstreams available")
	}

	// If we have a cached working base, try it first (fast path)
	if cached := GetWorkingBase(); cached != "" {
		procPath := strings.TrimLeft(procedure, "/")
		targetBase := strings.TrimRight(cached, "/") + "/" + procPath
		result, tryErr := callTRPCOnce(ctx, client, targetBase, requestBody)
		if tryErr == nil {
			return result, nil
		}
		// Cache miss - base went stale, clear it
		SetWorkingBase("")
	}

	// Race all remaining bases in parallel
	type baseResult struct {
		result UpstreamCallResult
		base   string
		err    error
	}
	ch := make(chan baseResult, len(bases))
	for _, base := range bases {
		go func(b string) {
			procPath := strings.TrimLeft(procedure, "/")
			targetBase := strings.TrimRight(b, "/") + "/" + procPath
			r, err := callTRPCOnce(ctx, client, targetBase, requestBody)
			ch <- baseResult{result: r, base: b, err: err}
		}(base)
	}

	remaining := len(bases)
	for remaining > 0 {
		br := <-ch
		remaining--
		if br.err == nil {
			SetWorkingBase(br.base)
			br.result.BaseURL = br.base
			// Drain remaining
			go func() {
				for remaining > 0 {
					<-ch
					remaining--
				}
			}()
			return br.result, nil
		}
		lastErr = br.err
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no TypeScript control-plane upstreams available")
	}
	return UpstreamCallResult{}, lastErr
}

// callTRPCOnce attempts a tRPC call: POST first, GET on 405.
func callTRPCOnce(ctx context.Context, client *http.Client, targetBase string, requestBody []byte) (UpstreamCallResult, error) {
	postReq, err := http.NewRequestWithContext(ctx, http.MethodPost, targetBase, bytes.NewReader(requestBody))
	if err != nil {
		return UpstreamCallResult{}, err
	}
	postReq.Header.Set("content-type", "application/json")

	resp, err := client.Do(postReq)
	if err != nil {
		return UpstreamCallResult{}, err
	}
	body, readErr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if readErr != nil {
		return UpstreamCallResult{}, readErr
	}

	if resp.StatusCode == 405 {
		inputVal := url.QueryEscape(fmt.Sprintf(`{"0":%s}`, string(requestBody)))
		getURL := targetBase + "?batch=1&input=" + inputVal
		getReq, getErr := http.NewRequestWithContext(ctx, http.MethodGet, getURL, nil)
		if getErr != nil {
			return UpstreamCallResult{}, getErr
		}
		getReq.Header.Set("accept", "application/json")
		resp, err = client.Do(getReq)
		if err != nil {
			return UpstreamCallResult{}, err
		}
		body, readErr = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return UpstreamCallResult{}, readErr
		}
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return UpstreamCallResult{}, fmt.Errorf("upstream %s returned %d: %s", targetBase, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	data, extractErr := extractTRPCData(body)
	if extractErr != nil {
		return UpstreamCallResult{}, extractErr
	}
	return UpstreamCallResult{Data: data}, nil
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
