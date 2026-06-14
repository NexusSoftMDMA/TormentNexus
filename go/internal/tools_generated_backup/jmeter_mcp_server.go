package tools

import "context"

func HandleRunTest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	testPlan, _ :=getString(args, "testPlan")
	_ = testPlan
	return success("JMeter test started: " + testPlan)
}

func HandleGetStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	testName, _ :=getString(args, "testName")
	_ = testName
	return ok("JMeter test status is unknown")
}