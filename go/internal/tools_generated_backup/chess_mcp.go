package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetChessRating(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := fmt.Sprintf("https://api.chess.com/pub/player/%s", username)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch player: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("player not found")
}

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response")
}

	rating, found := data["rating"]
	if !found {
		return err("rating not available")
}

	return ok(fmt.Sprintf("Rating: %v", rating))
}

func HandleGetChessPuzzle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://lichess.org/api/puzzle/random")
	if e != nil {
		return err("failed to fetch puzzle: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("puzzle service unavailable")
}

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse puzzle")
}

	puzzle, found := data["puzzle"]
	if !found {
		return err("no puzzle found")
}

	return success(fmt.Sprintf("Puzzle: %v", puzzle))
}