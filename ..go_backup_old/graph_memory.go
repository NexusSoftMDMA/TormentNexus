package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func HandleAddTriple(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subject, _ :=getString(args, "subject")
	predicate, _ :=getString(args, "predicate")
	object, _ :=getString(args, "object")
	if subject == "" || predicate == "" || object == "" {
		return err("subject, predicate, and object are required")
}

	triple := map[string]string{"subject": subject, "predicate": predicate, "object": object}
	body, _ := json.Marshal(triple)
	baseURL := os.Getenv("GRAPH_MEMORY_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	resp, e := http.DefaultClient.Post(baseURL+"/triples", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to add triple: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err("unexpected status: " + resp.Status)
}

	return ok("triple added successfully")
}