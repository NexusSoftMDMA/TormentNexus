package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// DownstreamDiscoveryContext provides context for downstream discovery operations.
type DownstreamDiscoveryContext struct {
	NamespaceUUID          string
	SessionID              string
	IncludeInactiveServers bool
}

// DownstreamServerVisit represents a visit to a downstream MCP server during discovery.
type DownstreamServerVisit struct {
	UUID       string
	ServerName string
	Client     *StdioClient
}

// DownstreamDiscovery handles discovery of tools, prompts, and resources from downstream MCP servers.
type DownstreamDiscovery struct {
	mu               sync.RWMutex
	promptToClient   map[string]*StdioClient
	resourceToClient map[string]*StdioClient
	discoveryTimeout time.Duration
}

// NewDownstreamDiscovery creates a new downstream discovery service.
func NewDownstreamDiscovery() *DownstreamDiscovery {
	return &DownstreamDiscovery{
		promptToClient:   make(map[string]*StdioClient),
		resourceToClient: make(map[string]*StdioClient),
		discoveryTimeout: 5 * time.Second,
	}
}

// SetDiscoveryTimeout sets the timeout for downstream discovery operations.
func (dd *DownstreamDiscovery) SetDiscoveryTimeout(timeout time.Duration) {
	dd.discoveryTimeout = timeout
}

// isSameServerInstance checks if a server is the same TormentNexus instance.
func isSameServerInstance(params map[string]interface{}, namespaceUUID string) bool {
	name, ok := params["name"].(string)
	if !ok {
		return false
	}
	return name == "tormentnexus-unified-"+namespaceUUID
}

// DiscoverTools discovers tools from all eligible downstream servers.
func (dd *DownstreamDiscovery) DiscoverTools(ctx context.Context, discCtx *DownstreamDiscoveryContext) ([]ToolSchema, error) {
	servers := dd.getEligibleServers(discCtx)
	var allTools []ToolSchema
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, visit := range servers {
		wg.Add(1)
		go func(v DownstreamServerVisit) {
			defer wg.Done()

			subCtx, cancel := context.WithTimeout(ctx, dd.discoveryTimeout)
			defer cancel()

			tools, err := dd.discoverToolsFromServer(subCtx, &v)
			if err != nil {
				fmt.Printf("[DownstreamDiscovery] Error discovering tools from %s: %v\n", v.ServerName, err)
				return
			}

			mu.Lock()
			allTools = append(allTools, tools...)
			mu.Unlock()
		}(visit)
	}

	wg.Wait()
	return allTools, nil
}

// DiscoverPrompts discovers prompts from all eligible downstream servers.
func (dd *DownstreamDiscovery) DiscoverPrompts(ctx context.Context, discCtx *DownstreamDiscoveryContext) ([]Prompt, error) {
	servers := dd.getEligibleServers(discCtx)
	var allPrompts []Prompt
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, visit := range servers {
		wg.Add(1)
		go func(v DownstreamServerVisit) {
			defer wg.Done()

			subCtx, cancel := context.WithTimeout(ctx, dd.discoveryTimeout)
			defer cancel()

			prompts, err := dd.discoverPromptsFromServer(subCtx, &v)
			if err != nil {
				fmt.Printf("[DownstreamDiscovery] Error discovering prompts from %s: %v\n", v.ServerName, err)
				return
			}

			mu.Lock()
			for _, p := range prompts {
				namespacedName := v.ServerName + "__" + p.Name
				p.Name = namespacedName
				dd.promptToClient[namespacedName] = v.Client
				allPrompts = append(allPrompts, p)
			}
			mu.Unlock()
		}(visit)
	}

	wg.Wait()
	return allPrompts, nil
}

// DiscoverResources discovers resources from all eligible downstream servers.
func (dd *DownstreamDiscovery) DiscoverResources(ctx context.Context, discCtx *DownstreamDiscoveryContext) ([]Resource, error) {
	servers := dd.getEligibleServers(discCtx)
	var allResources []Resource
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, visit := range servers {
		wg.Add(1)
		go func(v DownstreamServerVisit) {
			defer wg.Done()

			subCtx, cancel := context.WithTimeout(ctx, dd.discoveryTimeout)
			defer cancel()

			resources, err := dd.discoverResourcesFromServer(subCtx, &v)
			if err != nil {
				fmt.Printf("[DownstreamDiscovery] Error discovering resources from %s: %v\n", v.ServerName, err)
				return
			}

			mu.Lock()
			for _, r := range resources {
				dd.resourceToClient[r.URI] = v.Client
				allResources = append(allResources, r)
			}
			mu.Unlock()
		}(visit)
	}

	wg.Wait()
	return allResources, nil
}

func (dd *DownstreamDiscovery) getEligibleServers(discCtx *DownstreamDiscoveryContext) []DownstreamServerVisit {
	dd.mu.RLock()
	defer dd.mu.RUnlock()

	// In a full implementation, this would query the MCP server pool.
	// For now, return an empty list.
	return nil
}

// Prompt represents an MCP prompt.
type Prompt struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Resource represents an MCP resource.
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

func (dd *DownstreamDiscovery) discoverToolsFromServer(ctx context.Context, visit *DownstreamServerVisit) ([]ToolSchema, error) {
	if visit.Client == nil {
		return nil, fmt.Errorf("no client for server %s", visit.ServerName)
	}

	resp, err := visit.Client.Call(ctx, "tools/list", nil)
	if err != nil {
		return nil, fmt.Errorf("tools/list failed for %s: %w", visit.ServerName, err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tools/list error for %s: %v", visit.ServerName, resp.Error)
	}

	if resp.Result == nil {
		return nil, nil
	}

	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, err
	}

	var listResult struct {
		Tools []struct {
			Name        string      `json:"name"`
			Description string      `json:"description"`
			InputSchema interface{} `json:"inputSchema"`
		} `json:"tools"`
	}

	if err := json.Unmarshal(resultBytes, &listResult); err != nil {
		return nil, err
	}

	var tools []ToolSchema
	for _, t := range listResult.Tools {
		tools = append(tools, ToolSchema{
			Name:        NamespaceToolName(visit.ServerName, t.Name),
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}
	return tools, nil
}

func (dd *DownstreamDiscovery) discoverPromptsFromServer(ctx context.Context, visit *DownstreamServerVisit) ([]Prompt, error) {
	if visit.Client == nil {
		return nil, fmt.Errorf("no client for server %s", visit.ServerName)
	}

	resp, err := visit.Client.Call(ctx, "prompts/list", nil)
	if err != nil {
		return nil, fmt.Errorf("prompts/list failed for %s: %w", visit.ServerName, err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("prompts/list error for %s: %v", visit.ServerName, resp.Error)
	}

	if resp.Result == nil {
		return nil, nil
	}

	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, err
	}

	var listResult struct {
		Prompts []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"prompts"`
	}

	if err := json.Unmarshal(resultBytes, &listResult); err != nil {
		return nil, err
	}

	var prompts []Prompt
	for _, p := range listResult.Prompts {
		prompts = append(prompts, Prompt{
			Name:        p.Name,
			Description: p.Description,
		})
	}
	return prompts, nil
}

func (dd *DownstreamDiscovery) discoverResourcesFromServer(ctx context.Context, visit *DownstreamServerVisit) ([]Resource, error) {
	if visit.Client == nil {
		return nil, fmt.Errorf("no client for server %s", visit.ServerName)
	}

	resp, err := visit.Client.Call(ctx, "resources/list", nil)
	if err != nil {
		return nil, fmt.Errorf("resources/list failed for %s: %w", visit.ServerName, err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("resources/list error for %s: %v", visit.ServerName, resp.Error)
	}

	if resp.Result == nil {
		return nil, nil
	}

	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, err
	}

	var listResult struct {
		Resources []struct {
			URI         string `json:"uri"`
			Name        string `json:"name"`
			Description string `json:"description"`
			MimeType    string `json:"mimeType"`
		} `json:"resources"`
	}

	if err := json.Unmarshal(resultBytes, &listResult); err != nil {
		return nil, err
	}

	var resources []Resource
	for _, r := range listResult.Resources {
		resources = append(resources, Resource{
			URI:         r.URI,
			Name:        r.Name,
			Description: r.Description,
			MimeType:    r.MimeType,
		})
	}
	return resources, nil
}
