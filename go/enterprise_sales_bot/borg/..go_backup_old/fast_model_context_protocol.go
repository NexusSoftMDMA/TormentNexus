package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "world"
	}
	msg := fmt.Sprintf("Hello, %s!", name)
	return ok(msg)
}

func HandleAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	sum := a + b
	msg := fmt.Sprintf("%d + %d = %d", a, b, sum)
	return ok(msg)
}