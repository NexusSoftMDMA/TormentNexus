package tools

import (
	"context"
	"encoding/json"
	"os"
	"sync"
)

var mu sync.Mutex

func HandleRemember(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" {
		return err("Key is required")
}

	mu.Lock()
	defer mu.Unlock()
	data := make(map[string]string)
	f, e := os.Open("memory.json")
	if e == nil {
		json.NewDecoder(f).Decode(&data)
		f.Close()

	data[key] = value
	f, e = os.Create("memory.json")
	if e != nil {
		return err("Failed to save memory")
}

	defer f.Close()
	json.NewEncoder(f).Encode(data)
	return ok("Memory saved")
}

}

func HandleRecall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("Key is required")
}

	mu.Lock()
	defer mu.Unlock()
	data := make(map[string]string)
	f, e := os.Open("memory.json")
	if e != nil {
		return err("No memory found")
}

	defer f.Close()
	json.NewDecoder(f).Decode(&data)
	value, found := data[key]
	if !found {
		return err("Key not found")
}

	return success(value)
}