package tools

import (
	"context"
	"strings"
)

func HandleResumeToJobdescriptionMatcher(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resume, _ :=getString(args, "resume")
	jobDesc, _ :=getString(args, "job_description")
	if resume == "" || jobDesc == "" {
		return err("resume and job_description are required")
	}
	score := computeMatch(resume, jobDesc)
	return success("Match score: " + string(rune(score)) + "/100")
}

func computeMatch(resume, job string) int {
	rWords := strings.Fields(strings.ToLower(resume))
	jWords := strings.Fields(strings.ToLower(job))
	matches := 0
	for _, r := range rWords {
		for _, j := range jWords {
			if r == j && len(r) > 3 {
				matches++
				break
			}
		}
	}
	if len(jWords) == 0 {
		return 0
	}
	score := (matches * 100) / len(jWords)
	if score > 100 {
		return 100
	}
	return score
}// touch 1781132140
