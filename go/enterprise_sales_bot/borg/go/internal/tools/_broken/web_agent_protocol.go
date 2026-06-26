package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// HandleRecordAction receives a browser action event (as recorded by the WAP Chrome extension)
// and persists it to the local data directory organized by date and task ID.
func HandleRecordAction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	taskID, _ :=getString(args, "task_id")
	actionType, _ :=getString(args, "action_type")
	actionTimestamp, _ :=getString(args, "action_timestamp")
	targetID, _ :=getString(args, "target_id")
	targetClass, _ :=getString(args, "target_class")
	targetHTML, _ :=getString(args, "target_html")
	pageHTML, _ :=getString(args, "page_html")
	allEvents, _ :=getString(args, "all_events")

	if taskID == "" {
		return err("task_id is required")
}

	dataRoot, _ :=getString(args, "data_dir")
	if dataRoot == "" {
		dataRoot = "data"
	}

	now := time.Now()
	dateFolder := now.Format("20060102")
	ts := now.Format("20060102_150405")
	folderPath := filepath.Join(dataRoot, dateFolder, taskID)

	mkdirErr := os.MkdirAll(folderPath, 0755)
	if mkdirErr != nil {
		return err(mkdirErr.Error())
}

	filename := fmt.Sprintf("summary_event_%s.json", ts)
	filePath := filepath.Join(folderPath, filename)

	event := map[string]interface{}{
		"taskId":          taskID,
		"type":            actionType,
		"actionTimestamp": actionTimestamp,
		"eventTarget": map[string]interface{}{
			"type":         actionType,
			"target":       targetHTML,
			"targetId":     targetID,
			"targetClass":  targetClass,
		},
		"allEvents":       allEvents,
		"pageHTMLContent":  pageHTML,
	}

	data, jsonErr := json.MarshalIndent(event, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	writeErr := os.WriteFile(filePath, data, 0644)
	if writeErr != nil {
		return err(writeErr.Error())
}

	return ok(fmt.Sprintf("Event received and saved as %s", filePath))
}

// HandleGenerateReplayList reads recorded action events from a data directory
// and produces either an exact or smart replay list JSON file.
func HandleGenerateReplayList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dataDirPath, _ :=getString(args, "data_dir_path")
	outputDirPath, _ :=getString(args, "output_dir_path")
	mode, _ :=getString(args, "mode")
	taskID, _ :=getString(args, "task_id")

	if dataDirPath == "" {
		return err("data_dir_path is required")
}

	if outputDirPath == "" {
		return err("output_dir_path is required")
}

	if mode == "" {
		mode = "exact"
	}
	if taskID == "" {
		return err("task_id is required")
}

	entries, readErr := os.ReadDir(dataDirPath)
	if readErr != nil {
		return err(readErr.Error())
}

	var actions []map[string]interface{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		fullPath := filepath.Join(dataDirPath, entry.Name())
		data, fileErr := os.ReadFile(fullPath)
		if fileErr != nil {
			continue
		}
		var event map[string]interface{}
		jsonErr := json.Unmarshal(data, &event)
		if jsonErr != nil {
			continue
		}
		actions = append(actions, event)

	if len(actions) == 0 {
		return err("no action events found in " + dataDirPath)
}

	mkdirErr := os.MkdirAll(outputDirPath, 0755)
	if mkdirErr != nil {
		return err(mkdirErr.Error())
}

	var replayList map[string]interface{}
	if mode == "smart" {
		replayList = map[string]interface{}{
			"taskId":   taskID,
			"mode":     "smart",
			"actions":  actions,
			"summary":  fmt.Sprintf("Smart replay list with %d actions for task %s", len(actions), taskID),
		}
	} else {
		replayList = map[string]interface{}{
			"taskId":   taskID,
			"mode":     "exact",
			"actions":  actions,
			"summary":  fmt.Sprintf("Exact replay list with %d actions for task %s", len(actions), taskID),
		}
	}

	var outputFileName string
	if mode == "smart" {
		outputFileName = fmt.Sprintf("wap_smart_replay_list_%s.json", taskID)
	} else {
		outputFileName = fmt.Sprintf("wap_exact_replay_list_%s.json", taskID)

	outputPath := filepath.Join(outputDirPath, outputFileName)
	resultData, jsonErr := json.MarshalIndent(replayList, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	writeErr := os.WriteFile(outputPath, resultData, 0644)
	if writeErr != nil {
		return err(writeErr.Error())
}

	return ok(fmt.Sprintf("Generated %s replay list at %s with %d actions", mode, outputPath, len(actions)))
}

}
}

// HandleReplayActions replays a stored replay list by sending each action to a WAP replay service endpoint.
func HandleReplayActions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	replayListPath, _ :=getString(args, "wap_replay_list")
	replayServiceURL, _ :=getString(args, "replay_service_url")
	maxConcurrent, _ :=getInt(args, "max_concurrent")

	if replayListPath == "" {
		return err("wap_replay_list is required")
}

	if replayServiceURL == "" {
		replayServiceURL = "http://localhost:4934"
	}
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}

	data, fileErr := os.ReadFile(replayListPath)
	if fileErr != nil {
		return err(fileErr.Error())
}

	var replayList map[string]interface{}
	jsonErr := json.Unmarshal(data, &replayList)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	actionsRaw, found := replayList["actions"].([]interface{})
	if !found {
		return err("invalid replay list: missing or invalid actions array")
}

	taskID, _ := replayList["taskId"].(string)
	mode, _ := replayList["mode"].(string)
	if mode == "" {
		mode = "exact"
	}

	client := http.DefaultClient
	successCount := 0
	failCount := 0
	var errors []string

	for i, actionRaw := range actionsRaw {
		action, found := actionRaw.(map[string]interface{})
		if !found {
			failCount++
			errors = append(errors, fmt.Sprintf("action %d: invalid format", i))
			continue
		}

		action["taskId"] = taskID
		action["replayMode"] = mode
		action["stepIndex"] = i

		payload, marshalErr := json.Marshal(action)
		if marshalErr != nil {
			failCount++
			errors = append(errors, fmt.Sprintf("action %d: marshal error: %s", i, marshalErr.Error()))
			continue
		}

		replayURL := replayServiceURL + "/replay-action"
		req, reqErr := http.NewRequestWithContext(ctx, "POST", replayURL, strings.NewReader(string(payload)))
		if reqErr != nil {
			failCount++
			errors = append(errors, fmt.Sprintf("action %d: request error: %s", i, reqErr.Error()))
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, httpErr := client.Do(req)
		if httpErr != nil {
			failCount++
			errors = append(errors, fmt.Sprintf("action %d: http error: %s", i, httpErr.Error()))
			continue
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			successCount++
		} else {
			failCount++
			errors = append(errors, fmt.Sprintf("action %d: status %d", i, resp.StatusCode))

	}

	result := fmt.Sprintf("Replay complete (%s mode): %d/%d actions succeeded, %d failed",
		mode, successCount, len(actionsRaw), failCount)
	if len(errors) > 0 {
		result += "\nErrors: " + strings.Join(errors, "; ")

	return ok(result)
}

}
}

// HandleConvertToMCPServer converts a recorded replay list into an MCP server configuration
// that can be reused by any agent or user.
func HandleConvertToMCPServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	taskID, _ :=getString(args, "task_id")
	replayListPath, _ :=getString(args, "replay_list_path")
	outputDir, _ :=getString(args, "output_dir")

	if taskID == "" {
		return err("task_id is required")
}

	if replayListPath == "" {
		return err("replay_list_path is required")
}

	if outputDir == "" {
		outputDir = "mcp_servers"
	}

	data, fileErr := os.ReadFile(replayListPath)
	if fileErr != nil {
		return err(fileErr.Error())
}

	var replayList map[string]interface{}
	jsonErr := json.Unmarshal(data, &replayList)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	actionsRaw, found := replayList["actions"].([]interface{})
	if !found {
		return err("invalid replay list: missing or invalid actions array")
}

	mode, _ := replayList["mode"].(string)
	if mode == "" {
		mode = "exact"
	}

	serverDir := filepath.Join(outputDir, taskID)
	mkdirErr := os.MkdirAll(serverDir, 0755)
	if mkdirErr != nil {
		return err(mkdirErr.Error())
}

	// Build tool definitions from actions
	var tools []map[string]interface{}
	for i, actionRaw := range actionsRaw {
		action, found := actionRaw.(map[string]interface{})
		if !found {
			continue
		}
		actionType, _ := action["type"].(string)
		if actionType == "" {
			actionType = "action"
		}
		toolName := fmt.Sprintf("step_%d_%s", i+1, actionType)
		description := fmt.Sprintf("Replay step %d: %s action for task %s", i+1, actionType, taskID)

		eventTarget, _ := action["eventTarget"].(map[string]interface{})
		parameters := map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
		if eventTarget != nil {
			props := map[string]interface{}{
				"target_id":    map[string]interface{}{"type": "string", "description": "Target element ID"},
				"target_class": map[string]interface{}{"type": "string", "description": "Target element class"},
			}
			if tid, _ := eventTarget["targetId"].(string); tid != "" {
				props["target_id"] = map[string]interface{}{"type": "string", "description": "Target element ID", "default": tid}
			}
			if tclass, _ := eventTarget["targetClass"].(string); tclass != "" {
				props["target_class"] = map[string]interface{}{"type": "string", "description": "Target element class", "default": tclass}
			}
			parameters["properties"] = props
		}

		tools = append(tools, map[string]interface{}{
			"name":        toolName,
			"description": description,
			"parameters":  parameters,
		})

	mcpServer := map[string]interface{}{
		"name":        fmt.Sprintf("wap-server-%s", taskID),
		"version":     "1.0.0",
		"description": fmt.Sprintf("WAP MCP Server for task %s (%s replay)", taskID, mode),
		"tools":       tools,
		"replay_list": replayListPath,
		"task_id":     taskID,
		"mode":         mode,
		"total_steps":  strconv.Itoa(len(tools)),
	}

	serverData, jsonErr := json.MarshalIndent(mcpServer, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	serverFilePath := filepath.Join(serverDir, fmt.Sprintf("mcp_server_%s.json", taskID))
	writeErr := os.WriteFile(serverFilePath, serverData, 0644)
	if writeErr != nil {
		return err(writeErr.Error())
}

	return ok(fmt.Sprintf("MCP server config generated at %s with %d tools for task %s (%s mode)",
}
		serverFilePath, len(tools), taskID, mode))

}

// HandleListRecordedSessions scans the data directory for recorded sessions and returns a summary.
func HandleListRecordedSessions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dataDir, _ :=getString(args, "data_dir")
	if dataDir == "" {
		dataDir = "data"
	}

	var sessions []map[string]interface{}

	entries, readErr := os.ReadDir(dataDir)
	if readErr != nil {
		return err(readErr.Error())
}

	for _, dateEntry := range entries {
		if !dateEntry.IsDir() {
			continue
		}
		datePath := filepath.Join(dataDir, dateEntry.Name())
		taskEntries, taskErr := os.ReadDir(datePath)
		if taskErr != nil {
			continue
		}
		for _, taskEntry := range taskEntries {
			if !taskEntry.IsDir() {
				continue
			}
			taskPath := filepath.Join(datePath, taskEntry.Name())
			files, fileErr := os.ReadDir(taskPath)
			if fileErr != nil {
				continue
			}
			eventCount := 0
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
					eventCount++
				}
			}
			sessions = append(sessions, map[string]interface{}{
				"date":        dateEntry.Name(),
				"task_id":     taskEntry.Name(),
				"event_count": eventCount,
				"path":        taskPath,
			})

	}

	if len(sessions) == 0 {
		return ok("No recorded sessions found in " + dataDir)
}

	result, _ := json.MarshalIndent(sessions, "", "  ")
	return ok(string(result))
}

}

// HandleQueryReplayService queries the WAP replay service for status or task information.
func HandleQueryReplayService(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serviceURL, _ :=getString(args, "service_url")
	endpoint, _ :=getString(args, "endpoint")
	taskID, _ :=getString(args, "task_id")

	if serviceURL == "" {
		serviceURL = "http://localhost:4934"
	}
	if endpoint == "" {
		endpoint = "/status"
	}

	parsedURL, parseErr := url.Parse(serviceURL)
	if parseErr != nil {
		return err(parseErr.Error())
}

	queryPath := parsedURL.Path + endpoint
	if taskID != "" {
		if strings.Contains(queryPath, "?") {
			queryPath += "&task_id=" + url.QueryEscape(taskID)
		} else {
			queryPath += "?task_id=" + url.QueryEscape(taskID)

	}
	parsedURL.Path = queryPath

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return err(httpErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("service returned status %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}
}