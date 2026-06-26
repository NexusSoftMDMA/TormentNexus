package tools

import (
    "context"
)

func HandleListPipelines(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("List of ZenML pipelines retrieved successfully")
}

func HandleGetPipelineRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    runID, _ :=getString(args, "run_id")
    return success("Pipeline run details for run ID: " + runID)
}