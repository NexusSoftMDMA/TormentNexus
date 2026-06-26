package tools

import (
	"context"
	"encoding/json"
)

func HandleGetDomains(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domains := []string{
		"Autonomous Agents", "Multi-Agent Systems", "Reinforcement Learning",
		"LLM Agents", "Tool-Use Agents", "Planning & Reasoning",
		"Agentic Workflows", "Benchmarks & Evaluation", "Safety & Alignment",
		"Memory & Context", "Agentic Frameworks", "Code Agents",
		"Web Agents", "Robotics Agents", "Simulation Environments",
		"Agentic RAG", "Agentic Security", "Observability & Monitoring",
		"Agentic Testing", "Agentic UI/UX", "Agentic Economics",
		"Agentic Data", "Agentic Infrastructure", "Agentic Governance",
	}
	data, e := json.Marshal(domains)
	if e != nil {
		return err("failed to marshal domains")
}

	return success(string(data))
}

func HandleSearchResources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	query, _ :=getString(args, "query")
	resources := []map[string]string{
		{"title": "A Survey of Agentic AI", "domain": domain, "type": "paper"},
		{"title": "AgentBench: Evaluating LLMs as Agents", "domain": domain, "type": "benchmark"},
	}
	if query != "" {
		resources = append(resources, map[string]string{
			"title": "Agentic frameworks for " + query,
			"domain": domain,
			"type":   "framework",
		})

	data, e := json.Marshal(resources)
	if e != nil {
		return err("failed to marshal resources")
}

	return success(string(data))
}
}