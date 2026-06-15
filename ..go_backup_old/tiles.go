package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleFetchTile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	zoom, _ :=getInt(args, "zoom")
	x, _ :=getInt(args, "x")
	y, _ :=getInt(args, "y")
	url := fmt.Sprintf("https://tile.openstreetmap.org/%d/%d/%d.png", zoom, x, y)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("fetch error: " + e.Error())
}

	defer resp.Body.Close()
	_, e = io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("HTTP %d", resp.StatusCode))
}

	return ok("Tile fetched successfully")
}