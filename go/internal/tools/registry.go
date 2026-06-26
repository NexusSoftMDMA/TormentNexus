package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type ToolResponse struct {
	Content []TextContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func ok(text string) (ToolResponse, error) {
	return ToolResponse{
		Content: []TextContent{{Type: "text", Text: text}},
	}, nil
}

func err(msg string) (ToolResponse, error) {
	return ToolResponse{}, fmt.Errorf("%s", msg)
}

func getString(args map[string]interface{}, key string) (string, bool) {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

func getInt(args map[string]interface{}, key string) (int, bool) {
	if v, ok := args[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val), true
		case int:
			return val, true
		}
	}
	return 0, false
}

func getBool(args map[string]interface{}, key string) (bool, bool) {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b, true
		}
	}
	return false, false
}

type ToolHandler func(ctx context.Context, args map[string]interface{}) (ToolResponse, error)

type Registry struct {
	mu       sync.RWMutex
	handlers map[string]ToolHandler
}

func NewRegistry() *Registry {
	r := &Registry{
		handlers: make(map[string]ToolHandler),
	}
	r.Register("run_dag", HandleRunDAG)
	r.Register("read_file", HandleReadFile)
	r.Register("write_file", HandleWriteFile)
	r.Register("list_dir", HandleListDirectory)
	r.Register("delete_file", HandleDeleteFile)
	r.Register("ripgrep", HandleRipgrep)
	r.Register("search_text", HandleRipgrep)
	r.Register("search_web", HandleSearch)
	r.Register("list", HandleList)
	r.Register("info", HandleInfo)
	r.Register("version", HandleVersion)
	r.Register("semgrep_version", HandleSemgrepVersion)
	r.Register("semgrep_scan", HandleSemgrepScan)
	r.Register("execute_query", HandleExecuteQuery)
	r.Register("mem0", HandleMem0)
	r.Register("mem1", HandleMem1)
	r.Register("mem2", HandleMem2)
	r.Register("probe", HandleProbe)
	r.Register("code_research", HandleCodeResearch)
	r.Register("search_semantic", HandleSearchSemantic)
	r.Register("search_regex", HandleSearchRegex)
	r.Register("fetch", HandleFetch)
	r.Register("get", HandleGet)
	r.Register("post", HandlePost)
	r.Register("browser_action", HandleBrowserAction)
	r.Register("evolve", HandleEvolve)
	return r
}

func (r *Registry) Register(name string, handler ToolHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[name] = handler
}

func (r *Registry) Execute(ctx context.Context, name string, args map[string]interface{}) (ToolResponse, error) {
	r.mu.RLock()
	handler, ok := r.handlers[name]
	r.mu.RUnlock()
	if !ok {
		return ToolResponse{}, fmt.Errorf("unknown tool: %s", name)
	}
	return handler(ctx, args)
}

func (r *Registry) HasTool(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.handlers[name]
	return ok
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var names []string
	for name := range r.handlers {
		names = append(names, name)
	}
	return names
}

var _ = json.Marshal
