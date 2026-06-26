package mcpimpl

import (
	"context"
)

func HandleSubmitSurvey(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	answer, _ :=getString(args, "answer")
	if name == "" || answer == "" {
		return err("name and answer are required")
}

	return success("survey response submitted for " + name)
}

func HandleGetSurveyResults(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	surveyId, _ :=getString(args, "surveyId")
	if surveyId == "" {
		return err("surveyId is required")
}

	return ok("Results for survey " + surveyId + ": 5 responses")
}