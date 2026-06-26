package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func HandleLookupDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	typeName, _ :=getString(args, "typeName")
	if typeName == "" {
		return err("typeName is required")
}

	reqURL := "https://api.functype.dev/doc?type=" + url.QueryEscape(typeName)
	resp, e := http.DefaultClient.Get(reqURL)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}

func HandleValidateCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	payload := map[string]string{"code": code}
	b, e := json.Marshal(payload)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	reqURL := "https://api.functype.dev/validate"
	resp, e := http.DefaultClient.Post(reqURL, "application/json", strings.NewReader(string(b)))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}