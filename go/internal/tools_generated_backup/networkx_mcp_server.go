package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// HandleAnalyzeGraph analyzes a graph and returns node and edge counts.
func HandleAnalyzeGraph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	graphJSON, _ :=getString(args, "graph")
	if graphJSON == "" {
		return err("graph argument is required")
}

	var graph struct {
		Nodes []string   `json:"nodes"`
		Edges [][]string `json:"edges"`
	}
	if e := json.Unmarshal([]byte(graphJSON), &graph); e != nil {
		return err("invalid graph JSON: " + e.Error())
}

	nodeSet := make(map[string]struct{})
	for _, n := range graph.Nodes {
		nodeSet[n] = struct{}{}
	}
	for _, edge := range graph.Edges {
		if len(edge) >= 2 {
			nodeSet[edge[0]] = struct{}{}
			nodeSet[edge[1]] = struct{}{}
		}
	}
	nodeCount := len(nodeSet)
	edgeCount := len(graph.Edges)
	result := fmt.Sprintf("Nodes: %d, Edges: %d", nodeCount, edgeCount)
	return ok(result)
}