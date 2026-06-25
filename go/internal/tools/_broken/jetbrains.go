package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	cachedEndpoint string
	previousList   string
	client = http.DefaultClient
)

func getHost() string {
	h := os.Getenv("HOST")
	if h == "" {
		return "127.0.0.1"
	}
	return h
}

func logMsg(format string, args ...interface{}) {
	if os.Getenv("LOG_ENABLED") == "true" {
		fmt.Fprintf(os.Stderr, format+"\n", args...)

}

}

func testListTools(endpoint string) (bool, error) {
	logMsg("Sending test request to %s/mcp/list_tools", endpoint)
	resp, fetchErr := client.Get(endpoint + "/mcp/list_tools")
	if fetchErr != nil {
		logMsg("Error during testListTools for endpoint %s: %v", endpoint, fetchErr)
		return false, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logMsg("Test request to %s/mcp/list_tools failed with status %d", endpoint, resp.StatusCode)
		return false, nil
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		logMsg("Error reading response body: %v", readErr)
		return false, nil
	}

	currentResponse := string(body)
	logMsg("Received response from %s/mcp/list_tools: %s...", endpoint, truncateStr(currentResponse, 100))

	if previousList != "" && previousList != currentResponse {
		logMsg("Response has changed since the last check.")

	previousList = currentResponse

	return true, nil
}

}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func findWorkingIDEEndpoint() (string, error) {
	logMsg("Attempting to find a working IDE endpoint...")

	host := getHost()

	if p := os.Getenv("IDE_PORT"); p != "" {
		logMsg("IDE_PORT is set to %s. Testing this port.", p)
		testEndpoint := fmt.Sprintf("http://%s:%s/api", host, p)
		okVal, _ := testListTools(testEndpoint)
		if okVal {
			logMsg("IDE_PORT %s is working.", p)
			return testEndpoint, nil
		}
		return "", fmt.Errorf("specified IDE_PORT=%s but it is not responding correctly", p)
}

	if cachedEndpoint != "" {
		okVal, _ := testListTools(cachedEndpoint)
		if okVal {
			logMsg("Using cached endpoint, it's still working")
			return cachedEndpoint, nil
		}
	}

	for port := 63342; port <= 63352; port++ {
		candidateEndpoint := fmt.Sprintf("http://%s:%d/api", host, port)
		logMsg("Testing port %d...", port)
		okVal, _ := testListTools(candidateEndpoint)
		if okVal {
			logMsg("Found working IDE endpoint at %s", candidateEndpoint)
			return candidateEndpoint, nil
		}
		logMsg("Port %d is not responding correctly.", port)

	previousList = ""
	return "", fmt.Errorf("no working IDE endpoint found in range 63342-63352")
}

}

func updateIDEEndpoint() {
	endpoint, findErr := findWorkingIDEEndpoint()
	if findErr != nil {
		logMsg("Failed to update IDE endpoint: %v", findErr)
		return
	}
	cachedEndpoint = endpoint
	logMsg("Updated cachedEndpoint to: %s", cachedEndpoint)

}

func startEndpointUpdater() {
	updateIDEEndpoint()
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			updateIDEEndpoint()

	}()

}
}

func init() {
	startEndpointUpdater()

}

// HandleListTools lists available tools from the JetBrains IDE
func HandleListTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	logMsg("Handling list_tools request.")
	if cachedEndpoint == "" {
		return err("no working IDE endpoint available")
	}

	logMsg("Using cached endpoint %s to list tools.", cachedEndpoint)
	resp, fetchErr := client.Get(cachedEndpoint + "/mcp/list_tools")
	if fetchErr != nil {
		logMsg("Error fetching tools: %v", fetchErr)
		return err(fmt.Sprintf("unable to list tools: %v", fetchErr))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logMsg("Failed to fetch tools with status %d", resp.StatusCode)
		return err("unable to list tools")
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("error reading response: %v", readErr))
	}

	logMsg("Successfully fetched tools: %s", truncateStr(string(body), 200))
	return ok(string(body))
}

// HandleCallTool calls a specific tool on the JetBrains IDE
func HandleCallTool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	toolName, _ :=getString(args, "tool_name")
	if toolName == "" {
		return err("tool_name is required")
	}

	// Extract tool arguments - everything except tool_name
	toolArgs := make(map[string]interface{})
	for k, v := range args {
		if k != "tool_name" {
			toolArgs[k] = v
		}
	}

	logMsg("Handling tool call: name=%s, args=%v", toolName, toolArgs)
	if cachedEndpoint == "" {
		return err("no working IDE endpoint available")
	}

	logMsg("ENDPOINT: %s | Tool name: %s | args: %v", cachedEndpoint, toolName, toolArgs)

	bodyBytes, marshalErr := json.Marshal(toolArgs)
	if marshalErr != nil {
		return err(fmt.Sprintf("error marshaling arguments: %v", marshalErr))
	}

	endpoint := fmt.Sprintf("%s/mcp/%s", cachedEndpoint, toolName)
	req, reqErr := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(bodyBytes)))
	if reqErr != nil {
		return err(fmt.Sprintf("error creating request: %v", reqErr))
	}
	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		logMsg("Error in handleToolCall: %v", fetchErr)
		return err(fmt.Sprintf("request failed: %v", fetchErr))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logMsg("Response failed with status %d for tool %s", resp.StatusCode, toolName)
		return err(fmt.Sprintf("response failed: %d", resp.StatusCode))
	}

	var ideResp struct {
		Status string `json:"status"`
		Error  string `json:"error"`
	}
	decodeErr := json.NewDecoder(resp.Body).Decode(&ideResp)
	if decodeErr != nil {
		return err(fmt.Sprintf("error parsing IDE response: %v", decodeErr))
	}

	logMsg("Parsed response: status=%s error=%s", ideResp.Status, ideResp.Error)

	isError := ideResp.Error != ""
	text := ideResp.Status
	if isError {
		text = ideResp.Error
	}

	logMsg("Final response text: %s, isError: %v", text, isError)

	if isError {
		return err(text)
	}
	return ok(text)
}

// HandleCheckConnection checks if a JetBrains IDE is available and returns the endpoint info
func HandleCheckConnection(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host := getHost()
	port := os.Getenv("IDE_PORT")

	if cachedEndpoint == "" {
		if port != "" {
			return ok(fmt.Sprintf("No active IDE connection. Configured IDE_PORT=%s on host %s, but endpoint is not responding.", port, host))
		}
		return ok(fmt.Sprintf("No active IDE connection. Scanned ports 63342-63352 on host %s, but none responded.", host))
	}

	var portInfo string
	if port != "" {
		portInfo = fmt.Sprintf(" (configured IDE_PORT=%s)", port)

	return ok(fmt.Sprintf("Connected to JetBrains IDE at %s%s", cachedEndpoint, portInfo))
}

}

// HandleSetIDEPort manually sets the IDE port to connect to
func HandleSetIDEPort(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	portStr, _ :=getString(args, "port")
	if portStr == "" {
		return err("port is required")
	}

	portNum, convErr := strconv.Atoi(portStr)
	if convErr != nil {
		return err(fmt.Sprintf("invalid port number: %v", convErr))
	}
	if portNum < 1 || portNum > 65535 {
		return err("port must be between 1 and 65535")
	}

	os.Setenv("IDE_PORT", portStr)
	host := getHost()
	testEndpoint := fmt.Sprintf("http://%s:%s/api", host, portStr)

	okResult, testErr := testListTools(testEndpoint)
	if testErr != nil {
		return err(fmt.Sprintf("error testing endpoint: %v", testErr))
	}

	if okResult {
		cachedEndpoint = testEndpoint
		return ok(fmt.Sprintf("Successfully connected to IDE at %s", testEndpoint))
	}

	return ok(fmt.Sprintf("IDE_PORT set to %s, but endpoint %s is not responding correctly", portStr, testEndpoint))
}