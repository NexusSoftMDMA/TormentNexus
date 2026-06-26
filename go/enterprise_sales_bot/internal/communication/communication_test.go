package communication

import (
	"context"
	"testing"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestIntentClassifier_Classify(t *testing.T) {
	classifier := &MockIntentClassifier{}
	intent, err := classifier.Classify(context.Background(), "How much does HyperNexus cost?")
	if err != nil {
		t.Fatalf("Classification failed: %v", err)
	}
	if intent != IntentPricing {
		t.Errorf("Expected IntentPricing, got %s", intent)
	}
}

func TestSalesEngine_Decide(t *testing.T) {
	engine := NewLearningSalesEngine(nil, nil, nil)

	t.Run("Enterprise Pricing Intent", func(t *testing.T) {
		salesCtx := SalesContext{
			LatestIntent: IntentPricing,
			Company: db.Company{
				MarketCapTier: "Enterprise",
			},
			Deal: db.Deal{
				ID:           1,
				CurrentState: db.StateResearched,
			},
		}

		action, err := engine.Decide(context.Background(), salesCtx)
		if err != nil {
			t.Fatalf("Decision failed: %v", err)
		}
		if action != ActionRespond {
			t.Errorf("Expected ActionRespond for Enterprise pricing, got %s", action)
		}
	})

	t.Run("Should Advance State", func(t *testing.T) {
		salesCtx := SalesContext{
			LatestIntent: IntentPricing,
			Interactions: []db.Interaction{
				{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4},
			},
			Deal: db.Deal{
				ID:           2,
				CurrentState: db.StateEngaged,
			},
		}

		action, err := engine.Decide(context.Background(), salesCtx)
		if err != nil {
			t.Fatalf("Decision failed: %v", err)
		}
		if action != ActionAdvanceState {
			t.Errorf("Expected ActionAdvanceState when interest is high, got %s", action)
		}
	})
}
