package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func zabbixCall(url string, req map[string]interface{}) (map[string]interface{}, error) {
	body, e := json.Marshal(req)
	if e != nil {
		return nil, e
	}
	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(body))
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return nil, e
	}
	return result, nil
}