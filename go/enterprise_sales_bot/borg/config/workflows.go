package config

import (
	"encoding/json"
	"os"
)

type Workflow struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
}

var workflowsFile = "./.supercli/workflows.json"

func SaveWorkflow(w Workflow) error {
	var workflows []Workflow
	data, _ := os.ReadFile(workflowsFile)
	json.Unmarshal(data, &workflows)

	workflows = append(workflows, w)

	newData, _ := json.MarshalIndent(workflows, "", "  ")
	return os.WriteFile(workflowsFile, newData, 0644)
}

func ListWorkflows() []Workflow {
	var workflows []Workflow
	data, _ := os.ReadFile(workflowsFile)
	json.Unmarshal(data, &workflows)
	return workflows
}
