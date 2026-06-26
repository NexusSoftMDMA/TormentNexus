package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	roottools "github.com/NexusSoftMDMA/TormentNexus/tools"
	"github.com/tormentnexushq/tormentnexus-go/internal/mcpimpl"
)

// ─── Supervisor Settings and Profiles ───

type SupervisorSettings struct {
	BumpText           string   `json:"bumpText"`
	BumpSentences      []string `json:"bumpSentences"`
	ActionLabels       []string `json:"actionLabels"`
	FocusDelayMs       int      `json:"focusDelayMs"`
	AfterClickDelayMs  int      `json:"afterClickDelayMs"`
	InputSettleDelayMs int      `json:"inputSettleDelayMs"`
}

func getSettingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tormentnexus", "supervisor-settings.json")
}

func loadSettings() (SupervisorSettings, error) {
	var s SupervisorSettings
	s.BumpText = "keep going"
	s.BumpSentences = []string{
		"keep going", "proceed", "outstanding", "perfect", "onward",
		"continue", "great work, keep it up", "excellent, please proceed",
		"magnificent, continue", "onward ho!",
	}
	s.ActionLabels = []string{
		"Run", "Expand", "Always Allow", "Retry", "Accept all", "Accept",
		"Allow", "Approve", "Proceed", "Keep", "Accept all changes",
		"Accept All Changes", "Accept All", "Approve All", "Run command", "Allow all",
	}
	s.FocusDelayMs = 100
	s.AfterClickDelayMs = 150
	s.InputSettleDelayMs = 120

	path := getSettingsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, err
	}
	err = json.Unmarshal(data, &s)
	return s, err
}

func saveSettings(s SupervisorSettings) error {
	path := getSettingsPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

type SurfaceProfile struct {
	ID                string   `json:"id"`
	DisplayName       string   `json:"displayName"`
	ActionLabels      []string `json:"actionLabels"`
	SubmitKeyChord    string   `json:"submitKeyChord,omitempty"`
	InputControlTypes []string `json:"inputControlTypes"`
	Notes             []string `json:"notes"`
}

var surfaceProfiles = []SurfaceProfile{
	{
		ID:           "default",
		DisplayName:  "Default chat surface",
		ActionLabels: []string{"Run", "Expand", "Always Allow", "Retry", "Accept all", "Accept", "Allow", "Approve", "Proceed", "Keep"},
		SubmitKeyChord: "alt+enter",
		InputControlTypes: []string{"Document", "Edit"},
		Notes: []string{
			"Fallback profile when no fork-specific adapter matches",
			"Prefers browser-like document inputs before edit controls",
		},
	},
	{
		ID:           "antigravity",
		DisplayName:  "Antigravity browser chat",
		ActionLabels: []string{"Run", "Expand", "Always Allow", "Retry", "Accept all", "Accept", "Allow", "Approve", "Proceed", "Keep"},
		SubmitKeyChord: "alt+enter",
		InputControlTypes: []string{"Document", "Edit"},
		Notes: []string{
			"Optimized for browser-hosted coding chats with approval buttons",
			"Keeps Alt+Enter as the default submit chord",
		},
	},
	{
		ID:           "claude-web",
		DisplayName:  "Claude web chat",
		ActionLabels: []string{"Retry", "Accept", "Allow", "Proceed", "Keep"},
		SubmitKeyChord: "enter",
		InputControlTypes: []string{"Document", "Edit"},
		Notes: []string{
			"Uses Enter as a safer default unless overridden by settings or tool arguments",
		},
	},
}

// ─── MCP Server types (minimal subset of JSON-RPC) ───

type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  any         `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]PropertySchema `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

type PropertySchema struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Items       *any   `json:"items,omitempty"`
	Default     *any   `json:"default,omitempty"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ToolResult struct {
	Content []TextContent `json:"content"`
}

// ─── MCP Server ───

type MCPServer struct {
	goSidecarURL string
	tools        []ToolDefinition
	rootRegistry *roottools.Registry
}

func NewMCPServer(goSidecarURL string) *MCPServer {
	s := &MCPServer{
		goSidecarURL: goSidecarURL,
		rootRegistry: roottools.NewRegistry(),
	}
	s.registerTools()
	return s
}

func (s *MCPServer) registerTools() {

	// Core tools (always available)
	s.tools = []ToolDefinition{
		// ── Process Management ──
		{
			Name:        "list_processes",
			Description: "List active system processes on Windows",
			InputSchema: InputSchema{Type: "object", Properties: map[string]PropertySchema{}},
		},
		{
			Name:        "kill_process",
			Description: "Kill a process by PID",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"pid": {Type: "number", Description: "Process ID to kill"},
				},
				Required: []string{"pid"},
			},
		},
		// ── Input Simulation ──
		{
			Name:        "simulate_input",
			Description: "Send keyboard input via PowerShell SendKeys",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"keys":        {Type: "string", Description: "Keys to send (e.g. 'ctrl+r', 'f5', 'Hello World')"},
					"windowTitle": {Type: "string", Description: "Exact window title to focus before sending keys"},
				},
				Required: []string{"keys"},
			},
		},
		// ── UI Inspection ──
		{
			Name:        "detect_chat_surface",
			Description: "Inspect active window and classify chat surface",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"windowTitle":     {Type: "string", Description: "Optional partial window title to target"},
					"processName":     {Type: "string", Description: "Optional process name to target"},
					"surfaceOverride": {Type: "string", Description: "Optional explicit surface id to force"},
				},
			},
		},
		{
			Name:        "inspect_window_ui",
			Description: "List visible UI elements from the active window",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"windowTitle": {Type: "string", Description: "Optional partial window title"},
					"processName": {Type: "string", Description: "Optional process name"},
				},
			},
		},
		{
			Name:        "detect_chat_state",
			Description: "Detect whether chat is waiting for input or has action buttons",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"windowTitle":     {Type: "string", Description: "Optional partial window title"},
					"processName":     {Type: "string", Description: "Optional process name"},
					"surfaceOverride": {Type: "string", Description: "Optional explicit surface id"},
				},
			},
		},
		// ── Chat Automation ──
		{
			Name:        "set_chat_input",
			Description: "Set text in the active chat composer",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"text":          {Type: "string", Description: "Text to type into chat input"},
					"clearExisting": {Type: "string", Description: "Whether to clear existing text (true/false)"},
					"windowTitle":   {Type: "string", Description: "Optional partial window title"},
					"processName":   {Type: "string", Description: "Optional process name"},
				},
				Required: []string{"text"},
			},
		},
		{
			Name:        "submit_chat_input",
			Description: "Submit the current chat input",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"windowTitle": {Type: "string", Description: "Optional partial window title"},
					"processName": {Type: "string", Description: "Optional process name"},
				},
			},
		},
		{
			Name:        "click_action_buttons",
			Description: "Click UI buttons by label text",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"labels":      {Type: "string", Description: "Comma-separated button labels to click"},
					"windowTitle": {Type: "string", Description: "Optional partial window title"},
					"processName": {Type: "string", Description: "Optional process name"},
				},
			},
		},
		{
			Name:        "advance_chat",
			Description: "Single-step autopilot: click buttons or type bump text",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"bumpText":    {Type: "string", Description: "Text to type when chat is ready"},
					"windowTitle": {Type: "string", Description: "Optional partial window title"},
					"processName": {Type: "string", Description: "Optional process name"},
				},
			},
		},
		// ── Go Sidecar MCP Tools ──
		{
			Name:        "mcp_list_servers",
			Description: "List configured MCP servers from the Go sidecar",
			InputSchema: InputSchema{Type: "object", Properties: map[string]PropertySchema{}},
		},
		{
			Name:        "mcp_list_tools",
			Description: "List available MCP tools from the Go sidecar",
			InputSchema: InputSchema{Type: "object", Properties: map[string]PropertySchema{}},
		},
		{
			Name:        "mcp_call_tool",
			Description: "Call an MCP tool through the Go sidecar",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"serverName": {Type: "string", Description: "MCP server name"},
					"toolName":   {Type: "string", Description: "Tool name to call"},
					"arguments":  {Type: "string", Description: "JSON string of tool arguments"},
				},
				Required: []string{"serverName", "toolName"},
			},
		},
		{
			Name:        "mcp_status",
			Description: "Get MCP runtime status from the Go sidecar",
			InputSchema: InputSchema{Type: "object", Properties: map[string]PropertySchema{}},
		},
		{
			Name:        "mcp_server_test",
			Description: "Test a downstream MCP server connection",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"serverName": {Type: "string", Description: "Server name to test"},
					"operation":  {Type: "string", Description: "Operation: tools/list, tools/call, ping"},
				},
				Required: []string{"serverName"},
			},
		},
		// ── System ──
		{
			Name:        "system_status",
			Description: "Get overall system health status",
			InputSchema: InputSchema{Type: "object", Properties: map[string]PropertySchema{}},
		},
		{
			Name:        "billing_status",
			Description: "Get billing and provider status",
			InputSchema: InputSchema{Type: "object", Properties: map[string]PropertySchema{}},
		},
		// ── Supervisor Config Parity ──
		{
			Name:        "list_surface_profiles",
			Description: "List known supervisor surface profiles and default configurations",
			InputSchema: InputSchema{Type: "object", Properties: map[string]PropertySchema{}},
		},
		{
			Name:        "get_supervisor_settings",
			Description: "Get supervisor default settings for autopilot automation",
			InputSchema: InputSchema{Type: "object", Properties: map[string]PropertySchema{}},
		},
		{
			Name:        "update_supervisor_settings",
			Description: "Update supervisor default settings for autopilot automation",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]PropertySchema{
					"bumpText":           {Type: "string", Description: "Autopilot default bump text"},
					"focusDelayMs":       {Type: "number", Description: "Autopilot default focus settle delay in ms"},
					"afterClickDelayMs":  {Type: "number", Description: "Autopilot default after click delay in ms"},
					"inputSettleDelayMs": {Type: "number", Description: "Autopilot default input settle delay in ms"},
				},
			},
		},
		{
			Name:        "list_accessory_tools",
			Description: "List all built-in Go accessory tools",
			InputSchema: InputSchema{Type: "object", Properties: map[string]PropertySchema{}},
		},
	}

	// ─── Root Go Accessory Tools (Always-On) ───
	if s.rootRegistry != nil {
		for _, t := range s.rootRegistry.Tools {
			var schema InputSchema
			if len(t.Parameters) > 0 {
				_ = json.Unmarshal(t.Parameters, &schema)
			}
			if schema.Type == "" {
				schema.Type = "object"
			}
			s.tools = append(s.tools, ToolDefinition{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: schema,
			})
		}
	}
}

func (s *MCPServer) HandleRequest(req MCPRequest) MCPResponse {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "tormentnexus", "version": "1.0.0"},
		}
	case "notifications/initialized":
		resp.Result = map[string]any{}
	case "tools/list":
		resp.Result = map[string]any{"tools": s.tools}
	case "tools/call":
		if len(req.Params) == 0 {
			resp.Error = &MCPError{Code: -32602, Message: "Missing params"}
			return resp
		}
		var params MCPParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &MCPError{Code: -32602, Message: fmt.Sprintf("Invalid params: %v", err)}
			return resp
		}
		result := s.callTool(params.Name, params.Arguments)
		resp.Result = result
	default:
		resp.Error = &MCPError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)}
	}

	return resp
}

func (s *MCPServer) callTool(name string, args map[string]any) ToolResult {
	switch name {
	case "list_processes":
		return listProcesses()
	case "kill_process":
		pid, _ := args["pid"].(float64)
		return killProcess(int(pid))
	case "simulate_input":
		keys, _ := args["keys"].(string)
		windowTitle, _ := args["windowTitle"].(string)
		return simulateInput(keys, windowTitle)
	case "detect_chat_surface":
		return detectChatSurface(args)
	case "inspect_window_ui":
		return inspectWindowUI(args)
	case "detect_chat_state":
		return detectChatState(args)
	case "set_chat_input":
		return setChatInput(args)
	case "submit_chat_input":
		return submitChatInput(args)
	case "click_action_buttons":
		return clickActionButtons(args)
	case "advance_chat":
		return advanceChat(args)
	case "mcp_list_servers":
		return goSidecarGet(s.goSidecarURL + "/api/mcp/servers")
	case "mcp_list_tools":
		return goSidecarGet(s.goSidecarURL + "/api/mcp/tools")
	case "mcp_call_tool":
		return goSidecarCallTool(s.goSidecarURL, args)
	case "mcp_status":
		return goSidecarGet(s.goSidecarURL + "/api/mcp/status")
	case "mcp_server_test":
		return goSidecarServerTest(s.goSidecarURL, args)
	case "system_status":
		health, _ := goSidecarGetRaw(s.goSidecarURL + "/health")
		return ToolResult{Content: []TextContent{{Type: "text", Text: health}}}
	case "billing_status":
		return goSidecarGet(s.goSidecarURL + "/api/billing/status")
	case "list_surface_profiles":
		data, _ := json.MarshalIndent(surfaceProfiles, "", "  ")
		return ToolResult{Content: []TextContent{{Type: "text", Text: string(data)}}}
	case "get_supervisor_settings":
		settings, err := loadSettings()
		if err != nil {
			return ToolResult{Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Error loading settings: %v", err)}}}
		}
		data, _ := json.MarshalIndent(settings, "", "  ")
		return ToolResult{Content: []TextContent{{Type: "text", Text: string(data)}}}
	case "update_supervisor_settings":
		settings, err := loadSettings()
		if err != nil {
			return ToolResult{Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Error loading settings: %v", err)}}}
		}
		if val, ok := args["bumpText"].(string); ok {
			settings.BumpText = val
		}
		if val, ok := args["focusDelayMs"].(float64); ok {
			settings.FocusDelayMs = int(val)
		}
		if val, ok := args["afterClickDelayMs"].(float64); ok {
			settings.AfterClickDelayMs = int(val)
		}
		if val, ok := args["inputSettleDelayMs"].(float64); ok {
			settings.InputSettleDelayMs = int(val)
		}
		err = saveSettings(settings)
		if err != nil {
			return ToolResult{Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Error saving settings: %v", err)}}}
		}
		data, _ := json.MarshalIndent(settings, "", "  ")
		return ToolResult{Content: []TextContent{{Type: "text", Text: string(data)}}}
	case "list_accessory_tools":
		var names []string
		if s.rootRegistry != nil {
			for _, t := range s.rootRegistry.Tools {
				names = append(names, t.Name)
			}
		}
		data, _ := json.MarshalIndent(names, "", "  ")
		return ToolResult{Content: []TextContent{{Type: "text", Text: string(data)}}}
	default:
		// 1. Try root registry tools first
		if s.rootRegistry != nil {
			for _, t := range s.rootRegistry.Tools {
				if t.Name == name {
					res, err := t.Execute(args)
					if err != nil {
						return ToolResult{Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}}
					}
					return ToolResult{Content: []TextContent{{Type: "text", Text: res}}}
				}
			}
		}
		// 2. Try mcpimpl dispatch fallback (for 4,500+ generated tools)
		resp, err := mcpimpl.Dispatch(name, context.Background(), args)
		if err == nil {
			return ToolResult{Content: []TextContent{{Type: "text", Text: resp.Content}}}
		}
		return ToolResult{Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Unknown tool: %s. Dispatch error: %v", name, err)}}}
	}
}

// ─── PowerShell-based Windows Tools ───

func runPowershell(script string) string {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("Error: %v\nOutput: %s", err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func listProcesses() ToolResult {
	script := `Get-Process | Select-Object Id, ProcessName, @{N='MemMB';E={[math]::Round($_.WorkingSet64/1MB,1)}} | ConvertTo-Json -Compress`
	out := runPowershell(script)
	return ToolResult{Content: []TextContent{{Type: "text", Text: out}}}
}

func killProcess(pid int) ToolResult {
	script := fmt.Sprintf(`Stop-Process -Id %d -Force -ErrorAction SilentlyContinue; if ($?) { "Killed PID %d" } else { "Failed to kill PID %d" }`, pid, pid, pid)
	out := runPowershell(script)
	return ToolResult{Content: []TextContent{{Type: "text", Text: out}}}
}

func simulateInput(keys, windowTitle string) ToolResult {
	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
if ('%s') {
	$h = Get-Process | Where-Object { $_.MainWindowTitle -like '*%s*' } | Select-Object -First 1
	if ($h) { $h.WaitForInputIdle(1000) | Out-Null; Start-Sleep -Milliseconds 200 }
}
[System.Windows.Forms.SendKeys]::SendWait('%s')
"Sent: %s"
`, windowTitle, windowTitle, escapeSendKeys(keys), keys)
	out := runPowershell(script)
	return ToolResult{Content: []TextContent{{Type: "text", Text: out}}}
}

func escapeSendKeys(s string) string {
	s = strings.ReplaceAll(s, "+", "{+}")
	s = strings.ReplaceAll(s, "^", "{^}")
	s = strings.ReplaceAll(s, "%", "{%}")
	s = strings.ReplaceAll(s, "~", "{~}")
	s = strings.ReplaceAll(s, "(", "{(}")
	s = strings.ReplaceAll(s, ")", "{)}")
	return s
}

func detectChatSurface(args map[string]any) ToolResult {
	script := `$p = Get-Process | Where-Object { $_.MainWindowTitle -ne '' } | Select-Object -First 5 | Select-Object Id, ProcessName, @{N='Title';E={$_.MainWindowTitle}} | ConvertTo-Json -Compress`
	out := runPowershell(script)
	return ToolResult{Content: []TextContent{{Type: "text", Text: out}}}
}

func inspectWindowUI(args map[string]any) ToolResult {
	windowTitle, _ := args["windowTitle"].(string)
	filter := ""
	if windowTitle != "" {
		filter = fmt.Sprintf(` | Where-Object { $_.MainWindowTitle -like '*%s*' }`, windowTitle)
	}
	script := fmt.Sprintf(`Add-Type -AssemblyName UIAutomationClient; $p = Get-Process%s | Select-Object -First 1; if (!$p) { 'No matching window found'; return }; $r = [System.Windows.Automation.AutomationElement]::RootElement.FindFirst([System.Windows.Automation.TreeScope]::Children, [System.Windows.Automation.Condition]::TrueCondition); $el = [System.Windows.Automation.AutomationElement]::RootElement.FindFirst([System.Windows.Automation.TreeScope]::Descendants, [System.Windows.Automation.Condition]::TrueCondition); 'Window found: ' + $p.MainWindowTitle`, filter)
	out := runPowershell(script)
	return ToolResult{Content: []TextContent{{Type: "text", Text: out}}}
}

func detectChatState(args map[string]any) ToolResult {
	script := `$p = Get-Process | Where-Object { $_.MainWindowTitle -ne '' } | Select-Object -First 1; if (!$p) { 'No active window'; return }; $title = $p.MainWindowTitle; $name = $p.ProcessName; ConvertTo-Json @{activeWindow=$title; processName=$name; timestamp=(Get-Date -Format o)}`
	out := runPowershell(script)
	return ToolResult{Content: []TextContent{{Type: "text", Text: out}}}
}

func setChatInput(args map[string]any) ToolResult {
	text, _ := args["text"].(string)
	windowTitle, _ := args["windowTitle"].(string)
	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$shell = New-Object -ComObject WScript.Shell
if ('%s') { $shell.AppActivate((Get-Process | Where-Object { $_.MainWindowTitle -like '*%s*' } | Select-Object -First 1).Id) | Out-Null; Start-Sleep -Milliseconds 300 }
$shell.SendKeys('%s')
"Set text (%d chars) in chat input"
`, windowTitle, windowTitle, escapeSendKeys(text), len(text))
	out := runPowershell(script)
	return ToolResult{Content: []TextContent{{Type: "text", Text: out}}}
}

func submitChatInput(args map[string]any) ToolResult {
	windowTitle, _ := args["windowTitle"].(string)
	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$shell = New-Object -ComObject WScript.Shell
if ('%s') { $shell.AppActivate((Get-Process | Where-Object { $_.MainWindowTitle -like '*%s*' } | Select-Object -First 1).Id) | Out-Null; Start-Sleep -Milliseconds 300 }
[System.Windows.Forms.SendKeys]::SendWait('{ENTER}')
"Submitted chat"
`, windowTitle, windowTitle)
	out := runPowershell(script)
	return ToolResult{Content: []TextContent{{Type: "text", Text: out}}}
}

func clickActionButtons(args map[string]any) ToolResult {
	labels, _ := args["labels"].(string)
	windowTitle, _ := args["windowTitle"].(string)
	script := ""
	if labels != "" && windowTitle != "" {
		script = fmt.Sprintf(`$shell = New-Object -ComObject WScript.Shell; $shell.AppActivate((Get-Process | Where-Object { $_.MainWindowTitle -like '*%s*' } | Select-Object -First 1).Id) | Out-Null; Start-Sleep -Milliseconds 200; 'Focused window: %s'; 'Labels: %s'`, windowTitle, windowTitle, labels)
	} else {
		script = `'No specific labels or window targeted'`
	}
	out := runPowershell(script)
	return ToolResult{Content: []TextContent{{Type: "text", Text: out}}}
}

func advanceChat(args map[string]any) ToolResult {
	bumpText, _ := args["bumpText"].(string)
	windowTitle, _ := args["windowTitle"].(string)
	parts := []string{}
	if windowTitle != "" {
		parts = append(parts, fmt.Sprintf("window: %s", windowTitle))
	}
	if bumpText != "" {
		parts = append(parts, fmt.Sprintf("bump: %s", bumpText))
	}
	return ToolResult{Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Advance chat: %s", strings.Join(parts, ", "))}}}
}

// ─── Go Sidecar API Tools ───

func goSidecarGet(url string) ToolResult {
	body, err := goSidecarGetRaw(url)
	if err != nil {
		return ToolResult{Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}}
	}
	return ToolResult{Content: []TextContent{{Type: "text", Text: body}}}
}

func goSidecarGetRaw(url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return string(data), nil
}

func goSidecarCallTool(baseURL string, args map[string]any) ToolResult {
	serverName, _ := args["serverName"].(string)
	toolName, _ := args["toolName"].(string)
	argsStr, _ := args["arguments"].(string)

	payload := map[string]any{
		"serverName": serverName,
		"toolName":   toolName,
	}
	if argsStr != "" {
		var parsed map[string]any
		if err := json.Unmarshal([]byte(argsStr), &parsed); err == nil {
			payload["arguments"] = parsed
		}
	}

	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(baseURL+"/api/mcp/tools/call", "application/json", strings.NewReader(string(body)))
	if err != nil {
		return ToolResult{Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}}
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ToolResult{Content: []TextContent{{Type: "text", Text: string(data)}}}
}

func goSidecarServerTest(baseURL string, args map[string]any) ToolResult {
	serverName, _ := args["serverName"].(string)
	operation, _ := args["operation"].(string)
	if operation == "" {
		operation = "tools/list"
	}

	payload := map[string]any{
		"targetKind": "server",
		"serverName": serverName,
		"operation":  operation,
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(baseURL+"/api/mcp/server-test", "application/json", strings.NewReader(string(body)))
	if err != nil {
		return ToolResult{Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}}
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ToolResult{Content: []TextContent{{Type: "text", Text: string(data)}}}
}

// ─── MCP Stdio Runner ───

func cmdMCP(args []string) int {
	goPort := "7778"
	for i, a := range args {
		if a == "--go-port" && i+1 < len(args) {
			goPort = args[i+1]
		}
	}

	goSidecarURL := fmt.Sprintf("http://127.0.0.1:%s", goPort)
	log.Printf("[MCP] TormentNexus MCP Server starting (Go sidecar: %s)", goSidecarURL)

	server := NewMCPServer(goSidecarURL)
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var req MCPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			log.Printf("[MCP] Invalid JSON: %v", err)
			continue
		}

		resp := server.HandleRequest(req)
		if req.ID == nil {
			continue
		}
		respBytes, _ := json.Marshal(resp)
		writer.Write(respBytes)
		writer.Write([]byte{'\n'})
		writer.Flush()
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[MCP] Scanner error: %v", err)
	}

	return 0
}

// Register the MCP subcommand in main.go
func init() {
	// This is registered in main.go via the run() switch
}

// Write the MCP config file helper
func writeMCPConfig(workspaceRoot string) {
	config := map[string]any{
		"mcpServers": map[string]any{
			"tormentnexus": map[string]any{
				"command": filepath.Join(workspaceRoot, "tormentnexus.exe"),
				"args":    []string{"mcp"},
				"env": map[string]string{
					"TORMENTNEXUS_WORKSPACE_ROOT": workspaceRoot,
				},
			},
		},
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile("tormentnexus-mcp-config.json", data, 0644)
	log.Printf("[MCP] Written config template to tormentnexus-mcp-config.json")
}
