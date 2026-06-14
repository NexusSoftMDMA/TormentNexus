package tools

import (
	"context"
	"encoding/json"
)

func HandleGraph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	graph := map[string][]string{}
	if project != "" {
		graph["pipelineA"] = []string{"pipelineB", "pipelineC"}
		graph["pipelineB"] = []string{"pipelineD"}
	} else {
		graph["defaultPipeline"] = []string{}
	}
	data, e := json.Marshal(graph)
	if e != nil {
		return err("failed to serialize graph: " + e.Error())
}

	return success(string(data))
}