package communication

import (
	"context"
	"strings"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

func TestLLMResponseGenerator_Generate(t *testing.T) {
	provider := &llm.MockLLMProvider{}
	generator := NewRAGResponseGenerator(nil, provider)

	salesCtx := SalesContext{
		Contact: db.Contact{Name: "John Doe", Role: "CTO"},
		Company: db.Company{Name: "TechCorp"},
		Deal:    db.Deal{TechnicalDossier: "INFRASTRUCTURE_BOTTLENECK detected in legacy k8s clusters."},
		LatestIntent: IntentTechnical,
		Interactions: []db.Interaction{
			{RawText: "How can TormentNexus help with our scaling?"},
		},
	}

	resp, err := generator.Generate(context.Background(), salesCtx, ActionRespond)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	if !strings.Contains(resp, "John Doe") {
		t.Errorf("Response should contain contact name")
	}
	if !strings.Contains(resp, "INFRASTRUCTURE_BOTTLENECK") {
		t.Errorf("Response should contain technical dossier context")
	}
}
