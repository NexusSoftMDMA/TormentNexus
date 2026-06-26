package mcpimpl

import (
    "context"
    "fmt"
)

func HandleListExams(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    exams := []string{"AWS Certified Solutions Architect", "Azure Administrator", "Google Cloud Architect"}
    return success(fmt.Sprintf("Available exams: %v", exams))
}

func HandleGetExam(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    id, _ :=getString(args, "exam_id")
    if id == "" {
        return err("exam_id is required")
}

    details := fmt.Sprintf("Exam details for %s: 50 questions, 65 minutes, passing score 720", id)
    return success(details)
}