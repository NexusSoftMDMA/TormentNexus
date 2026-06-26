package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleDumpWindowHierarchy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	compressed := true
	if v, exists := args["compressed"]; exists {
		if b, found := v.(bool); found {
			compressed = b
		}
	}

	hierarchy := map[string]interface{}{
		"compressed": compressed,
		"root": map[string]interface{}{
			"class":    "android.widget.FrameLayout",
			"children": []interface{}{},
		},
	}
	jsonBytes, marshalErr := json.Marshal(hierarchy)
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal window hierarchy: %v", marshalErr))
}

	return ok(string(jsonBytes))
}

func HandleClick(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pointX, _ :=getInt(args, "pointX")
	pointY, _ :=getInt(args, "pointY")
	result := fmt.Sprintf("clicked at coordinates (%d, %d)", pointX, pointY)
	return ok(result)
}

func HandleSwipe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fromX, _ :=getInt(args, "fromX")
	fromY, _ :=getInt(args, "fromY")
	toX, _ :=getInt(args, "toX")
	toY, _ :=getInt(args, "toY")
	step := 32
	if v, exists := args["step"]; exists {
		if i, found := v.(int); found {
			step = i
		}
	}
	result := fmt.Sprintf("swiped from (%d,%d) to (%d,%d) with step %d", fromX, fromY, toX, toY, step)
	return ok(result)
}

func HandleGetDeviceInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deviceInfo := map[string]interface{}{
		"brand":           "Google",
		"model":           "Pixel 7",
		"screen_width":    1080,
		"screen_height":   2400,
		"android_version": "14",
		"is_screen_on":    true,
		"is_locked":       false,
	}
	jsonBytes, marshalErr := json.Marshal(deviceInfo)
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal device info: %v", marshalErr))
}

	return ok(string(jsonBytes))
}

func HandleShowToast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	result := fmt.Sprintf("toast displayed: %s", message)
	return ok(result)
}

func HandleGetprop(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	var propValue string
	switch name {
	case "ro.product.model":
		propValue = "Pixel 7"
	case "ro.build.version.release":
		propValue = "14"
	case "ro.product.brand":
		propValue = "Google"
	default:
		propValue = ""
	}
	return ok(propValue)
}