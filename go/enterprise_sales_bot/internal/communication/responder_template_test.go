package communication

import (
    "context"
    "testing"
    "github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

func TestGenerateFromTemplate(t *testing.T) {
    tmpl := &db.Template{
        ID:      "test-email",
        Subject: "Hello {{contact}} from {{company}}",
        Body:    "Hi {{contact}},\nWe noticed your work on {{tech_stack}} and think TormentNexus can help.",
        Channel: "email",
    }

    salesCtx := SalesContext{
        Company: db.Company{Name: "Acme Corp", TechStack: []string{"Go", "Kubernetes"}},
        Contact: db.Contact{Name: "Alice"},
    }

    rg := &RAGResponseGenerator{}
    subject, body, err := rg.GenerateFromTemplate(context.Background(), tmpl, salesCtx)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if subject != "Hello Alice from Acme Corp" {
        t.Errorf("subject mismatch: got %q", subject)
    }
    expectedBody := "Hi Alice,\nWe noticed your work on Go, Kubernetes and think TormentNexus can help."
    if body != expectedBody {
        t.Errorf("body mismatch: got %q expected %q", body, expectedBody)
    }
}
