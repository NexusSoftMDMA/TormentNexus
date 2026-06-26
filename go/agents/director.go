package agents

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/NexusSoftMDMA/TormentNexus/foundation/adapters"
	foundationorchestration "github.com/NexusSoftMDMA/TormentNexus/foundation/orchestration"
)

// Director Agent translates the TS core Director orchestrator.
// It acts as the primary task planner, coordinating sub-agents.
type Director struct {
	Name         string
	Provider     ILLMProvider
	State        map[string]interface{}
	History      []Message
	WorkingDir   string
	HyperAdapter *adapters.TormentNexusAdapter
}

func NewDirector(provider ILLMProvider) *Director {
	cwd, _ := os.Getwd()
	hyperAdapter := adapters.NewTormentNexusAdapter(cwd)
	return &Director{
		Name:         "Director",
		Provider:     provider,
		State:        make(map[string]interface{}),
		WorkingDir:   cwd,
		HyperAdapter: hyperAdapter,
		History: []Message{
			{
				Role:    RoleSystem,
				Content: strings.Join([]string{"You are the TormentNexus TechLead Director. Your role is absolute architectural supervision. Plan, delegate, and review.", hyperAdapter.BuildSystemContext()}, "\n\n"),
			},
		},
	}
}

func (d *Director) GetName() string {
	return d.Name
}

func (d *Director) GetRole() string {
	return "supervisor"
}

func (d *Director) HandleInput(ctx context.Context, input string) (string, error) {
	d.History = append(d.History, Message{Role: RoleUser, Content: input})

	plan, err := foundationorchestration.BuildPlan(foundationorchestration.PlanRequest{
		Prompt:     input,
		WorkingDir: d.WorkingDir,
	})
	if err == nil {
		d.State["lastPlan"] = plan
	}

	providerMessages := append([]Message(nil), d.History...)
	if err == nil {
		providerMessages = append(providerMessages, Message{
			Role:    RoleSystem,
			Content: strings.Join([]string{"Execution planning context:", plan.SystemContextHint, strings.Join(plan.Steps, "\n")}, "\n"),
		})
	}

	responseMsg, providerErr := d.Provider.Chat(ctx, providerMessages, []Tool{})
	if providerErr != nil {
		return "", fmt.Errorf("director execution failed: %w", providerErr)
	}

	d.History = append(d.History, responseMsg)
	if err == nil {
		return fmt.Sprintf("[Director Plan]\n- task type: %s\n- route: %s/%s\n\n%s", plan.TaskType, plan.Execution.Route.Provider, plan.Execution.Route.Model, responseMsg.Content), nil
	}
	return responseMsg.Content, nil
}

func (d *Director) InjectSystemContext(context string) {
	d.History[0].Content += "\n\n" + context
}

func (d *Director) GetState() map[string]interface{} {
	return d.State
}

// Example Stubs for other agents to achieve parity:

type Coder struct{ Director } // Inherits base logic for simplicity in this stub
type MetaArchitect struct{ Director }
type Researcher struct{ Director }
type Council struct{ Director }
