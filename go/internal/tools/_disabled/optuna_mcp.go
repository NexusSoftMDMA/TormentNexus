package tools

import (
	"context"
	"net/http"
)

func HandleOptimize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	studyName, _ :=getString(args, "study_name")
	nTrials, _ :=getInt(args, "n_trials")
	direction, _ :=getString(args, "direction")

	if nTrials == 0 {
		nTrials = 10
	}
	if direction == "" {
		direction = "minimize"
	}

	resp, e := http.DefaultClient.Get("https://example.com/optimize?study=" + studyName + "&trials=" + string(rune(nTrials)) + "&direction=" + direction)
	if e != nil {
		return err("failed to call optimize: " + e.Error())
}

	defer resp.Body.Close()

	return ok("optimization started for study '" + studyName + "' with " + string(rune(nTrials)) + " trials (" + direction + ")")
}

func HandleListStudies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://example.com/studies")
	if e != nil {
		return err("failed to list studies: " + e.Error())
}

	defer resp.Body.Close()

	return success("studies retrieved")
}