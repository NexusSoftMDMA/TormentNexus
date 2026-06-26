package main

import (
	"context"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

type CustomSource struct{}

func (s *CustomSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	return []db.Company{
		{Name: "PluginCompany", Domain: "plugin.com", TechStack: []string{"Go"}, HiringSignals: []string{"test"}},
	}, nil
}
func (s *CustomSource) HealthCheck() error { return nil }
func (s *CustomSource) Name() string { return "CustomPluginSource" }

var Source CustomSource
