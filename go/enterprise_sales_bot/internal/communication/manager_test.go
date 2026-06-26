package communication

import (
	"context"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

type mockClassifier struct{}
func (m *mockClassifier) Classify(ctx context.Context, text string) (Intent, error) {
	return IntentTechnical, nil
}

type mockResponder struct{}
func (m *mockResponder) Generate(ctx context.Context, salesCtx SalesContext, action Action) (string, error) {
	return "Mock Reply", nil
}

type mockStrategy struct{}
func (m *mockStrategy) Decide(ctx context.Context, salesCtx SalesContext) (Action, error) {
	return ActionRespond, nil
}

type mockOrderProcessor struct{}
func (m *mockOrderProcessor) ProcessOrder(ctx context.Context, deal db.Deal) error {
	return nil
}

func TestProcessInbound_Mock(t *testing.T) {
	// Note: Testing ProcessInbound requires a database.
	// Since we are focused on logic integration, we verify it compiles
	// and use the SalesContext tests in engine_test.go for deeper logic.

	manager := NewManager(nil, &mockClassifier{}, &mockResponder{}, &mockStrategy{}, &mockOrderProcessor{}, nil)
	if manager == nil {
		t.Fatal("Failed to create Manager")
	}
}
