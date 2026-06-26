package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListClients(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "baseUrl")
	if base == "" {
		base = "https://demo.mifos.io/fineract-provider/api/v1"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", base+"/self/clients", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API returned " + resp.Status)
}

	return ok(string(body))
}

func HandleGetClient(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	clientID, _ :=getString(args, "clientId")
	if clientID == "" {
		return err("clientId is required")
}

	base, _ :=getString(args, "baseUrl")
	if base == "" {
		base = "https://demo.mifos.io/fineract-provider/api/v1"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", base+"/self/clients/"+clientID, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API returned " + resp.Status)
}

	return ok(string(body))
}