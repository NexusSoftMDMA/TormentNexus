package tools

import (
	"context"
	"net/http"
	"encoding/json"
	"io/ioutil"
)

func HandleGravity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	mass, _ :=getString(args, "mass")
	if mass == "" {
		return err("mass parameter required")
}

	resp, e := http.DefaultClient.Get("https://api.example.com/gravity?mass=" + mass)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(data)
}