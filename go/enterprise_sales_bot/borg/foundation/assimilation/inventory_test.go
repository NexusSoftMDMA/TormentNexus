package assimilation

import "testing"

func TestInventoryCoversKeySources(t *testing.T) {
	items := Inventory()
	if len(items) < 20 {
		t.Fatalf("expected at least 20 source toolchains, got %d", len(items))
	}

	wanted := map[string]bool{
		"tormentnexus": false,
		"pi":        false,
		"aider":     false,
		"opencode":  false,
		"goose":     false,
	}

	for _, item := range items {
		if _, ok := wanted[item.ID]; ok {
			wanted[item.ID] = true
		}
	}

	for id, found := range wanted {
		if !found {
			t.Fatalf("expected source %q in inventory", id)
		}
	}
}

func TestCategoriesReturnsCounts(t *testing.T) {
	summary := Categories(Inventory())
	if len(summary) == 0 {
		t.Fatal("expected category summary entries")
	}
	if summary[0].Count < 1 {
		t.Fatalf("unexpected summary: %#v", summary[0])
	}
}
