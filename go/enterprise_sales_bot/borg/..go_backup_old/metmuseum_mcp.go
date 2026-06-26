package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func HandleSearchArtworks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	isOnView, _ :=getInt(args, "isOnView")
	url := "https://collectionapi.metmuseum.org/public/collection/v1/search?q=" + query + "&isOnView=" + strconv.Itoa(isOnView)
	resp, e := http.Get(url)
	if e != nil {
		return err("failed to search: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleGetObject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "objectID")
	url := "https://collectionapi.metmuseum.org/public/collection/v1/objects/" + strconv.Itoa(id)
	resp, e := http.Get(url)
	if e != nil {
		return err("failed to fetch object: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var obj map[string]interface{}
	json.Unmarshal(body, &obj)
	data, _ := json.Marshal(obj)
	return ok(string(data))
}