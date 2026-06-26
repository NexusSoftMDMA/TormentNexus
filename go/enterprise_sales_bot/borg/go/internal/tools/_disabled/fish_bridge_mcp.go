package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetFish(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	species, _ :=getString(args, "species")
	url := fmt.Sprintf("https://api.fishbridge.example/fish?species=%s", species)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch fish data: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}