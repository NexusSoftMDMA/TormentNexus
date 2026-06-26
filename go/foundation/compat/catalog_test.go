package compat

import "testing"

func TestDefaultCatalogIncludesPiToolContracts(t *testing.T) {
	catalog := DefaultCatalog()
	if catalog.Count() != 7 {
		t.Fatalf("expected 7 contracts, got %d", catalog.Count())
	}

	sources := catalog.Sources()
	if len(sources) != 1 || sources[0] != "pi" {
		t.Fatalf("unexpected sources: %#v", sources)
	}

	for _, name := range []string{"read", "write", "edit", "bash", "grep", "find", "ls"} {
		contracts := catalog.Lookup(name)
		if len(contracts) != 1 {
			t.Fatalf("expected one contract for %q, got %d", name, len(contracts))
		}
		if !contracts[0].ExactName || !contracts[0].ExactParameters || !contracts[0].ExactResultShape {
			t.Fatalf("contract for %q is not marked exact: %#v", name, contracts[0])
		}
	}
}

func TestCatalogRejectsMissingFields(t *testing.T) {
	catalog := NewCatalog()
	if err := catalog.Register(ToolContract{Name: "read"}); err == nil {
		t.Fatal("expected missing source error")
	}
	if err := catalog.Register(ToolContract{Source: "pi"}); err == nil {
		t.Fatal("expected missing name error")
	}
}
