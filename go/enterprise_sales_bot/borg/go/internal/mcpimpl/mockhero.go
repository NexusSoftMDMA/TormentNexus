package mcpimpl

import (
	"context"
	"encoding/json"
	"math/rand"
)

func HandleMockUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "John Doe"
	}
	age, _ :=getInt(args, "age")
	if age == 0 {
		age = 30
	}
	user := map[string]interface{}{
		"name": name,
		"age":  age,
	}
	data, e := json.Marshal(user)
	if e != nil {
		return err("failed to marshal user")
}

	return success(string(data))
}

func HandleMockRandom(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	min, _ :=getInt(args, "min")
	max, _ :=getInt(args, "max")
	if min >= max {
		min = 0
		max = 100
	}
	value := rand.Intn(max-min+1) + min
	return success(map[string]int{"random": value})
}