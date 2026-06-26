package tools

import (
	"context"
	"testing"
)

func BenchmarkRegistryExecute_SkillList(b *testing.B) {
	r := NewRegistry()
	ctx := context.Background()
	args := map[string]interface{}{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := r.Execute(ctx, "skill_list", args)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRegistryExecute_PromptList(b *testing.B) {
	r := NewRegistry()
	ctx := context.Background()
	args := map[string]interface{}{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := r.Execute(ctx, "prompt_list", args)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRegistryExecute_LS(b *testing.B) {
	r := NewRegistry()
	ctx := context.Background()
	args := map[string]interface{}{"path": "."}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := r.Execute(ctx, "ls", args)
		if err != nil {
			b.Fatal(err)
		}
	}
}
