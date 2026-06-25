package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	defaultHost       = "127.0.0.1"
	defaultPort       = 3025
	identitySignature = "mcp-browser-connector-24x7"
	discoveryTimeout  = 2 * time.Second
	apiTimeout        = 30 * time.Second
)

var (
	discoveredHost   string
	discoveredPort   int
	serverDiscovered bool
)

func getServerPort() int {
	if envPort := os.Getenv("BROWSER_TOOLS_PORT"); envPort != "" {
		if p, parseErr := strconv.Atoi(envPort); parseErr == nil && p > 0 {
			return p
		}
	}
	return defaultPort
}

func getServerHost() string {
	if envHost := os.Getenv("BROWSER_TOOLS_HOST"); envHost != "" {
		return envHost
	}
	return defaultHost
}

func discoverServer() bool {
	hosts := []string{getServerHost(), "127.0.0.1", "localhost"}
	seenPorts := map[int]bool{}
	ports := []int{}
	basePort := getServerPort()
	ports = append(ports, basePort)
	seenPorts[basePort] = true
	for p := 3025; p <= 3035; p++ {
		if !seenPorts[p] {
			ports = append(ports, p)
			seenPorts[p] = true
		}
	}

	client := http.DefaultClient

	for _, host := range hosts {
		for _, port := range ports {
			url := fmt.Sprintf("http://%s:%d/.identity", host, port)
			resp, reqErr := client.Get(url)
			if reqErr != nil {
				continue
			}
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				continue
			}
			var identity map[string]interface{}
			if jsonErr := json.Unmarshal(body, &identity); jsonErr != nil {
				continue
			}
			if sig, found := identity["signature"].(string); found && sig == identitySignature {
				discoveredHost = host
				discoveredPort = port
				serverDiscovered = true
				return true
			}
		}
	}
	return false
}

func withServerConnection(apiCall func() (string, error)) (ToolResponse, error) {
	if !serverDiscovered {
		if !discoverServer() {
			return err("Failed to discover browser connector server. Please ensure it's running.")
		}
	}
	result, apiErr := apiCall()
	if apiErr != nil {
		serverDiscovered = false
		if discoverServer() {
			result, retryErr := apiCall()
			if retryErr != nil {
				return err(fmt.Sprintf("Error after reconnection attempt: %s", retryErr.Error()))
			}
			return ok(result)
		}
		return err(fmt.Sprintf("Failed to reconnect to server: %s", apiErr.Error()))
	}
	return ok(result)
}

func fetchEndpoint(endpoint string) (string, error) {
	client := http.DefaultClient
	url := fmt.Sprintf("http://%s:%d%s", discoveredHost, discoveredPort, endpoint)
	resp, reqErr := client.Get(url)
	if reqErr != nil {
		return "", reqErr
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", readErr
	}
	return string(body), nil
}

func postEndpoint(endpoint string) (string, error) {
	client := http.DefaultClient
	url := fmt.Sprintf("http://%s:%d%s", discoveredHost, discoveredPort, endpoint)
	resp, reqErr := client.Post(url, "application/json", nil)
	if reqErr != nil {
		return "", reqErr
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", readErr
	}
	return string(body), nil
}

func formatJSON(raw string) string {
	var data interface{}
	if unmarshalErr := json.Unmarshal([]byte(raw), &data); unmarshalErr != nil {
		return raw
	}
	formatted, marshalErr := json.MarshalIndent(data, "", "  ")
	if marshalErr != nil {
		return raw
	}
	return string(formatted)
}

// HandleGetConsoleLogs retrieves browser console logs from the browser-tools-server
func HandleGetConsoleLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return withServerConnection(func() (string, error) {
		raw, fetchErr := fetchEndpoint("/console-logs")
		if fetchErr != nil {
			return "", fetchErr
		}
		return formatJSON(raw), nil
	})
}

// HandleGetConsoleErrors retrieves browser console errors from the browser-tools-server
func HandleGetConsoleErrors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return withServerConnection(func() (string, error) {
		raw, fetchErr := fetchEndpoint("/console-errors")
		if fetchErr != nil {
			return "", fetchErr
		}
		return formatJSON(raw), nil
	})
}

// HandleGetNetworkErrors retrieves network error logs from the browser-tools-server
func HandleGetNetworkErrors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return withServerConnection(func() (string, error) {
		raw, fetchErr := fetchEndpoint("/network-errors")
		if fetchErr != nil {
			return "", fetchErr
		}
		return formatJSON(raw), nil
	})
}

// HandleGetNetworkLogs retrieves all network logs from the browser-tools-server
func HandleGetNetworkLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return withServerConnection(func() (string, error) {
		raw, fetchErr := fetchEndpoint("/network-success")
		if fetchErr != nil {
			return "", fetchErr
		}
		return formatJSON(raw), nil
	})
}

// HandleTakeScreenshot captures a screenshot of the current browser tab
func HandleTakeScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return withServerConnection(func() (string, error) {
		raw, fetchErr := postEndpoint("/capture-screenshot")
		if fetchErr != nil {
			return "", fetchErr
		}
		return raw, nil
	})
}

// HandleGetSelectedElement retrieves the currently selected DOM element from the browser
func HandleGetSelectedElement(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return withServerConnection(func() (string, error) {
		raw, fetchErr := fetchEndpoint("/selected-element")
		if fetchErr != nil {
			return "", fetchErr
		}
		return formatJSON(raw), nil
	})
}