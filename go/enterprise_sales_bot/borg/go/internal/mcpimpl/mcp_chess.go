package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleGetLegalMoves(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fen, _ :=getString(args, "fen")
	if fen == "" {
		return err("missing fen")
}

	apiURL := fmt.Sprintf("https://lichess.org/api/board/legal-moves?fen=%s", url.QueryEscape(fen))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("api request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("json decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("legal moves: %v", result["moves"]))
}

func HandleMakeMove_mcp_chess(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fen, _ :=getString(args, "fen")
	move, _ :=getString(args, "move")
	if fen == "" || move == "" {
		return err("missing fen or move")
}

	apiURL := fmt.Sprintf("https://lichess.org/api/board/move?fen=%s&move=%s", url.QueryEscape(fen), url.QueryEscape(move))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("api request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("json decode failed: " + e.Error())
}

	return success(fmt.Sprintf("new fen: %v", result["fen"]))
}