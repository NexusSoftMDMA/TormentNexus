package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func HandleCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	pin, _ :=getInt(args, "pin")
	val, _ :=getInt(args, "value")
	host, _ :=getString(args, "host")
	if host == "" {
		host = "192.168.1.100"
	}
	url := fmt.Sprintf("http://%s/%s?pin=%d&val=%d", host, cmd, pin, val)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to send command: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleSensor(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pin, _ :=getInt(args, "pin")
	host, _ :=getString(args, "host")
	if host == "" {
		host = "192.168.1.100"
	}
	url := fmt.Sprintf("http://%s/sensor?pin=%d", host, pin)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("sensor request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read sensor data: " + e.Error())
}

	val, e := strconv.Atoi(string(body))
	if e != nil {
		return err("invalid sensor value: " + e.Error())
}

	return success(fmt.Sprintf("Sensor value: %d", val))
}