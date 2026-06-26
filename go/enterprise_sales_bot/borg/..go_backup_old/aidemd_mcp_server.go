package tools

import "context"

func HandleAIDEOverview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	overview := "AIDE methodology: Analyze, Iterate, Develop, Evaluate. " +
		"Use progressive disclosure to learn each step. Call AIDEStep with step=a/i/d/e for details."
	return ok(overview)
}

func HandleAIDEStep(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	step, _ :=getString(args, "step")
	var detail string
	if step == "a" {
		detail = "Analyze: Understand the problem, gather requirements, and define success criteria. " +
			"Start with 'why' before 'how'."
	} else if step == "i" {
		detail = "Iterate: Build in small cycles. Test early, gather feedback, and refine. " +
			"Each iteration improves understanding."
	} else if step == "d" {
		detail = "Develop: Implement the solution using chosen tools and patterns. " +
			"Write clean, modular code that is easy to maintain."
	} else if step == "e" {
		detail = "Evaluate: Validate against requirements, measure performance, and review. " +
			"Use data to decide if more iterations are needed."
	} else {
		detail = "Invalid step. Use: a (Analyze), i (Iterate), d (Develop), e (Evaluate)."
		return err(detail)
}

	return ok(detail)
}