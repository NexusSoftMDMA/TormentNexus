package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleNacosListServices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "nacosHost")
	if host == "" {
		host = "localhost"
	}
	port, _ :=getInt(args, "nacosPort")
	if port == 0 {
		port = 8848
	}
	url := fmt.Sprintf("http://%s:%d/nacos/v1/ns/service/list?pageNo=1&pageSize=100", host, port)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to request Nacos: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.Marshal(result)
	return success(string(data))
}

func HandleNacosGetInstance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "nacosHost")
	if host == "" {
		host = "localhost"
	}
	port, _ :=getInt(args, "nacosPort")
	if port == 0 {
		port = 8848
	}
	serviceName, _ :=getString(args, "serviceName")
	if serviceName == "" {
		return err("serviceName is required")
}

	url := fmt.Sprintf("http://%s:%d/nacos/v1/ns/instance/list?serviceName=%s", host, port, serviceName)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to request Nacos: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	data, _ := json.Marshal(result)
	return success(string(data))
}