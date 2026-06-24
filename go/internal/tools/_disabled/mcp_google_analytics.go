package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleReadReport(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "accessToken")
	propertyID, _ :=getString(args, "propertyId")
	bodyMap := map[string]interface{}{}
	for k, v := range args {
		if k == "accessToken" || k == "propertyId" {
			continue
		}
		bodyMap[k] = v
	}
	bodyBytes, e := json.Marshal(bodyMap)
	if e != nil {
		return err("failed to marshal request body: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("https://analyticsdata.googleapis.com/v1beta/properties/%s:runReport", propertyID),
		bytes.NewReader(bodyBytes))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("report result: %v", result))
}

func HandleSendEvent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	measurementID, _ :=getString(args, "measurementId")
	apiSecret, _ :=getString(args, "apiSecret")
	clientID, _ :=getString(args, "clientId")
	eventName, _ :=getString(args, "eventName")
	eventParamsStr, _ :=getString(args, "eventParams")
	var eventParams map[string]interface{}
	if eventParamsStr != "" {
		if e := json.Unmarshal([]byte(eventParamsStr), &eventParams); e != nil {
			return err("invalid eventParams JSON: " + e.Error())

	}
	payload := map[string]interface{}{
		"client_id": clientID,
		"events": []map[string]interface{}{
			{"name": eventName, "params": eventParams},
		},
	}
	bodyBytes, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	url := fmt.Sprintf("https://www.google-analytics.com/mp/collect?measurement_id=%s&api_secret=%s", measurementID, apiSecret)
	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return ok("event sent successfully")
}
}