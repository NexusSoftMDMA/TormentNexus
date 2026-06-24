package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleGetInvoice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	api, _ :=getString(args, "api_url")
	if api == "" {
		return err("api_url required")
}

	resp, e := http.DefaultClient.Get(api + "/invoice")
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success(fmt.Sprintf("%v", result))
}

func HandlePayInvoice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	api, _ :=getString(args, "api_url")
	invoice, _ :=getString(args, "invoice")
	if api == "" || invoice == "" {
		return err("api_url and invoice required")
}

	body := fmt.Sprintf(`{"invoice":"%s"}`, invoice)
	resp, e := http.DefaultClient.Post(api+"/pay", "application/json", strings.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return success(fmt.Sprintf("%v", result))
}