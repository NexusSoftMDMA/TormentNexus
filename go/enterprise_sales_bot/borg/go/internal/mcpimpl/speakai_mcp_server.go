package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleSearchRecordings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	baseURL := "https://api.speakai.com/v1"
	url := fmt.Sprintf("%s/recordings?query=%s", baseURL, query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to search recordings: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}

func HandleCreateClip(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	recordingID, _ :=getString(args, "recording_id")
	startTime, _ :=getString(args, "start_time")
	endTime, _ :=getString(args, "end_time")
	baseURL := "https://api.speakai.com/v1"
	url := fmt.Sprintf("%s/recordings/%s/clips", baseURL, recordingID)
	payload := map[string]interface{}{
		"start_time": startTime,
		"end_time":   endTime,
	}
	jsonPayload, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal payload: %v", e))
}

	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if e != nil {
		return err(fmt.Sprintf("failed to create clip: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}