//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleGetNearOptionPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	expiry, _ :=getInt(args, "expiry")
	active, _ :=getBool(args, "active")
	extra, found := args


-reasoner (deepseek)*,
}