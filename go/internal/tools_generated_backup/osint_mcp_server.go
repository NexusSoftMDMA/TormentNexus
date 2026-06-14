package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleShodan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ip, _ :=getString(args, "ip")
	key, _ :=getString(args, "api_key")
	if ip == "" || key == "" {
		return err("ip and api_key are required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.shodan.io/shodan/host/%s?key=%s", ip, key))
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	return ok(fmt.Sprintf("Shodan result: %v", result))
}

func HandleCrtSh(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("domain is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://crt.sh/?q=%s&output=json", domain))
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var certs []map[string]interface{}
	if e := json.Unmarshal(body, &certs); e != nil {
		return err(e.Error())
}

	return success(fmt.Sprintf("Found %d certificates", len(certs)))
}