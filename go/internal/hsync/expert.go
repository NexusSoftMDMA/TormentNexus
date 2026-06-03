package hsync

import (
	"context"
	"fmt"

	"github.com/tormentnexushq/tormentnexus-go/internal/mcp"
	"github.com/tormentnexushq/tormentnexus-go/internal/orchestration"
)

type ExpertManager struct {
	director  *orchestration.Director
	predictor *mcp.ToolPredictor
}

func NewExpertManager(director *orchestration.Director, predictor *mcp.ToolPredictor) *ExpertManager {
	return &ExpertManager{
		director:  director,
		predictor: predictor,
	}
}

type ExpertResult struct {
	Success bool   `json:"success"`
	TaskId  string `json:"taskId"`
}

func (m *ExpertManager) ExpertResearch(ctx context.Context, topic string) (ExpertResult, error) {
	fmt.Printf("[Go Expert] 🔍 Starting expert research for: %s\n", topic)
	
	goal := fmt.Sprintf("Deeply research the following topic and provide a comprehensive summary: %s", topic)
	err := m.director.StartAutonomousTask(ctx, goal)
	if err != nil {
		return ExpertResult{Success: false}, err
	}

	return ExpertResult{Success: true, TaskId: "task-research-go"}, nil
}

func (m *ExpertManager) ExpertCode(ctx context.Context, instruction string) (ExpertResult, error) {
	fmt.Printf("[Go Expert] 💻 Starting expert code generation for: %s\n", instruction)
	
	goal := fmt.Sprintf("Implement the following code instruction: %s", instruction)
	err := m.director.StartAutonomousTask(ctx, goal)
	if err != nil {
		return ExpertResult{Success: false}, err
	}

	return ExpertResult{Success: true, TaskId: "task-code-go"}, nil
}

func (m *ExpertManager) PredictTools(ctx context.Context, history string, goal string) ([]string, error) {
	return m.predictor.PredictAndPreload(ctx, history, goal)
}
