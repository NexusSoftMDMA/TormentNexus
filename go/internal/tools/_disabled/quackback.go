package tools

import (
	"context"
	"fmt"
)

func HandleGetQuack(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 1
	}
	s := "Quack!"
	for i := 1; i < count; i++ {
		s += " Quack!"
	}
	return ok(s)
}

func HandleGetDuckFact(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fact := "Ducks have waterproof feathers."
	return ok(fact)
}