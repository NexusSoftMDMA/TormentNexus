package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSendTriggeredSend(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accessToken, _ :=getString(args, "accessToken")
	baseURL, _ :=getString(args, "baseUrl")
	triggeredSendKey, _ :=getString(args, "triggeredSendKey")
	subscriberKey, _ :=getString(args, "subscriberKey")
	emailAddress, _ :=getString(args, "emailAddress")

	url := fmt.Sprintf("%s/triggeredSend/v1/definitions/%s/send", baseURL, triggeredSendKey)
	payload := map[string]interface{}{
		"To": map[string]string{
			"Address": emailAddress,
			"SubscriberKey": subscriberKey,
		},
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal payload: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(bytes.NewReader(body)))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("SFMC returned %d: %s", resp.StatusCode, string(respBody)))
}

	return success("Triggered send queued successfully")
}

func HandleGetDataExtensionRows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accessToken, _ :=getString(args, "accessToken")
	baseURL, _ :=getString(args, "baseUrl")
	key, _ :=getString(args, "dataExtensionKey")
	top, _ :=getInt(args, "top")

	url := fmt.Sprintf("%s/hub/v1/dataextensions/key:%s/rows?$top=%d", baseURL, key, top)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("SFMC returned %d: %s", resp.StatusCode, string(respBody)))
}

	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
}

	return ok(fmt.Sprintf("Data extension rows: %+v", result))
}