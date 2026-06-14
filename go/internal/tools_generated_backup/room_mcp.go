package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleCreateRoom(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	capacity, _ :=getInt(args, "capacity")
	if name == "" || capacity <= 0 {
		return err("name and capacity required")
	}
	apiURL := os.Getenv("ROOM_API_URL")
	if apiURL == "" {
		apiURL = "https://api.example.com/rooms"
	}
	body, e := json.Marshal(map[string]interface{}{"name": name, "capacity": capacity})
	if e != nil {
		return err("marshal failed")
	}
	resp, e := http.DefaultClient.Post(apiURL, "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed: "+e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("bad status")
	}
	var result struct{ ID string `json:"id"` }
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed")
	}
	return ok("Room created with id " + result.ID)
}

func HandleJoinRoom(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	roomID, _ :=getString(args, "room_id")
	user, _ :=getString(args, "user")
	if roomID == "" || user == "" {
		return err("room_id and user required")
	}
	apiURL := os.Getenv("ROOM_API_URL")
	if apiURL == "" {
		apiURL = "https://api.example.com/rooms"
	}
	body, e := json.Marshal(map[string]string{"room_id": roomID, "user": user})
	if e != nil {
		return err("marshal failed")
	}
	resp, e := http.DefaultClient.Post(apiURL+"/join", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed: "+e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("join failed")
	}
	return ok("Joined room successfully")
}