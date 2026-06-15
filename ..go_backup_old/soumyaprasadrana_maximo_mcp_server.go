package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	object, _ :=getString(args, "object")
	where, _ :=getString(args, "where")
	u, e := url.Parse(base + "/maximo/oslc/os/" + object)
	if e != nil {
		return err("invalid URL: " + e.Error())
}

	q := u.Query()
	q.Set("oslc.select", "*")
	if where != "" {
		q.Set("oslc.where", where)

	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err("request failed: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json parse error: " + e.Error())
}

	return ok(fmt.Sprintf("Query result: %+v", result))
}

}

func HandleStage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	object, _ :=getString(args, "object")
	action, _ :=getString(args, "action")
	data, _ :=getString(args, "data")
	u, e := url.Parse(base + "/maximo/oslc/os/" + object)
	if e != nil {
		return err("invalid URL: " + e.Error())
}

	q := u.Query()
	q.Set("_action", action)
	u.RawQuery = q.Encode()
	body := strings.NewReader(data)
	req, e := http.NewRequestWithContext(ctx, "POST", u.String(), body)
	if e != nil {
		return err("request failed: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	b, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(fmt.Sprintf("Stage result: %s", string(b)))
}