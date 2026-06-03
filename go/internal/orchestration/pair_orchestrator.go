package orchestration

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
)

type PairRole string

const (
	Planner     PairRole = "planner"
	Implementer PairRole = "implementer"
	Tester      PairRole = "tester"
	Critic      PairRole = "critic"
)

type SessionState string

const (
	StatePlanning    SessionState = "planning"
	StateReviewing   SessionState = "reviewing"
	StateRefining    SessionState = "refining"
	StateImplementing SessionState = "implementing"
	StateVerifying   SessionState = "verifying"
	StateEvaluating   SessionState = "evaluating"
	StateRevising    SessionState = "revising"
	StateCompleted   SessionState = "completed"
	StateFailed      SessionState = "failed"
)

type SquadMember struct {
	Name     string   `json:"name"`
	Role     PairRole `json:"role"`
	Provider string   `json:"provider"`
	ModelID  string   `json:"modelId"`
}

type PairSessionResult struct {
	Success     bool     `json:"success"`
	History     []string `json:"history"`
	FinalOutput string   `json:"finalOutput"`
	State       string   `json:"state"`
}

type PairOrchestrator struct {
	mu          sync.RWMutex
	Squad       []SquadMember
	History     []string
	State       SessionState
	Task        string
	CurrentRole PairRole
	Bus         interface {
		EmitEvent(eventType string, source string, payload interface{})
	}
	Consensus *ConsensusEngine
}

func NewPairOrchestrator(consensus *ConsensusEngine) *PairOrchestrator {
	return &PairOrchestrator{
		History:   []string{},
		State:     StatePlanning,
		Consensus: consensus,
	}
}

func (p *PairOrchestrator) SetEventBus(bus interface {
	EmitEvent(eventType string, source string, payload interface{})
}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Bus = bus
}

func (p *PairOrchestrator) SetupSquad(members []SquadMember) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Squad = members
}

func (p *PairOrchestrator) SetupFrontierSquad() {
	p.SetupSquad([]SquadMember{
		{Name: "Claude (Architect)", Role: Planner, Provider: "anthropic", ModelID: "claude-3-5-sonnet-20241022"},
		{Name: "GPT (Engineer)", Role: Implementer, Provider: "openai", ModelID: "gpt-4o"},
		{Name: "Gemini (Reviewer)", Role: Tester, Provider: "google", ModelID: "gemini-1.5-pro"},
		{Name: "Qwen (Auditor)", Role: Critic, Provider: "google", ModelID: "gemini-2.5-flash"},
	})
}

func (p *PairOrchestrator) RunTask(ctx context.Context, task string) (*PairSessionResult, error) {
	p.mu.Lock()
	p.Task = task
	p.History = []string{"USER: " + task}
	p.State = StatePlanning
	p.mu.Unlock()

	for {
		p.mu.RLock()
		state := p.State
		p.mu.RUnlock()

		switch state {
		case StatePlanning:
			plan, err := p.executeTurn(ctx, Planner, "Create a detailed implementation plan for this task: "+p.Task)
			if err != nil { return p.failSession(err) }
			p.addHistory(fmt.Sprintf("PLANNER (%s): %s", p.getMemberName(Planner), plan))
			p.transition(StateReviewing)

		case StateReviewing:
			lastEntry := p.getLastHistory()
			feedback, err := p.executeTurn(ctx, Tester, "Review this plan and identify potential edge cases or bugs: "+lastEntry)
			if err != nil { return p.failSession(err) }
			p.addHistory(fmt.Sprintf("TESTER (%s): %s", p.getMemberName(Tester), feedback))
			p.transition(StateRefining)

		case StateRefining:
			lastEntry := p.getLastHistory()
			finalPlan, err := p.executeTurn(ctx, Planner, "Refine the plan based on this feedback: "+lastEntry)
			if err != nil { return p.failSession(err) }
			p.addHistory(fmt.Sprintf("PLANNER (%s): %s", p.getMemberName(Planner), finalPlan))
			p.transition(StateImplementing)

		case StateImplementing:
			lastEntry := p.getLastHistory()
			implementation, err := p.executeTurn(ctx, Implementer, "Implement the final plan. Focus on correctness and performance. Plan: "+lastEntry)
			if err != nil { return p.failSession(err) }
			p.addHistory(fmt.Sprintf("IMPLEMENTER (%s): %s", p.getMemberName(Implementer), implementation))
			p.transition(StateVerifying)

		case StateVerifying:
			lastEntry := p.getLastHistory()
			verification, err := p.executeTurn(ctx, Tester, "Verify the implementation against the plan and task requirements. Implementation: "+lastEntry)
			if err != nil { return p.failSession(err) }
			p.addHistory(fmt.Sprintf("TESTER (%s): %s", p.getMemberName(Tester), verification))
			p.transition(StateEvaluating)

		case StateEvaluating:
			// Invoke Consensus Engine
			p.mu.RLock()
			models := []string{}
			for _, m := range p.Squad { models = append(models, m.Provider+"/"+m.ModelID) }
			p.mu.RUnlock()

			outcome, err := p.Consensus.Resolve(ctx, p.getLastHistory(), models)
			if err != nil { return p.failSession(err) }

			if outcome.Agreed {
				p.addHistory(fmt.Sprintf("CRITIC: Consensus reached. %s", outcome.Summary))
				p.transition(StateCompleted)
				return p.getResult(true), nil
			} else {
				p.addHistory(fmt.Sprintf("CRITIC: Rejecting implementation. Reason: %s", outcome.Summary))
				p.transition(StateRevising)
			}

		case StateRevising:
			lastRejection := p.getLastHistory()
			revision, err := p.executeTurn(ctx, Implementer, "Revise your implementation based on the CRITIC rejection: "+lastRejection)
			if err != nil { return p.failSession(err) }
			p.addHistory(fmt.Sprintf("IMPLEMENTER (%s) [REVISION]: %s", p.getMemberName(Implementer), revision))
			p.transition(StateVerifying)

		case StateCompleted, StateFailed:
			return p.getResult(state == StateCompleted), nil

		default:
			return nil, fmt.Errorf("unknown orchestrator state: %s", state)
		}
	}
}

func (p *PairOrchestrator) executeTurn(ctx context.Context, role PairRole, prompt string) (string, error) {
	p.mu.Lock()
	p.CurrentRole = role
	bus := p.Bus
	var member *SquadMember
	for i := range p.Squad {
		if p.Squad[i].Role == role {
			member = &p.Squad[i]
			break
		}
	}
	p.mu.Unlock()

	if member == nil { return "", fmt.Errorf("no member assigned to role: %s", role) }

	if bus != nil {
		bus.EmitEvent("swarm:turn_start", "PairOrchestrator", map[string]interface{}{
			"role": string(role),
			"name": member.Name,
			"model": member.ModelID,
			"prompt": prompt,
		})
	}

	systemPrompt := fmt.Sprintf(`You are part of a multi-agent squad. Name: %s. Role: %s.`, member.Name, strings.ToUpper(string(member.Role)))
	p.mu.RLock()
	turnPrompt := fmt.Sprintf("HISTORY:\n%s\n\nTURN (%s): %s", strings.Join(p.History, "\n"), strings.ToUpper(string(member.Role)), prompt)
	p.mu.RUnlock()

	resp, err := ai.AutoRouteWithModel(ctx, member.Provider+"/"+member.ModelID, []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: turnPrompt},
	})

	if err != nil { return "", err }

	if bus != nil {
		bus.EmitEvent("swarm:turn_end", "PairOrchestrator", map[string]interface{}{
			"role": string(role),
			"success": true,
			"content": resp.Content,
		})
	}
	return resp.Content, nil
}

func (p *PairOrchestrator) addHistory(entry string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.History = append(p.History, entry)
}

func (p *PairOrchestrator) getLastHistory() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.History) == 0 { return "" }
	return p.History[len(p.History)-1]
}

func (p *PairOrchestrator) transition(newState SessionState) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.State = newState
}

func (p *PairOrchestrator) getMemberName(role PairRole) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, m := range p.Squad {
		if m.Role == role { return m.Name }
	}
	return "Unknown"
}

func (p *PairOrchestrator) failSession(err error) (*PairSessionResult, error) {
	p.transition(StateFailed)
	p.addHistory(fmt.Sprintf("SYSTEM ERROR: %v", err))
	return p.getResult(false), err
}

func (p *PairOrchestrator) getResult(success bool) *PairSessionResult {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return &PairSessionResult{
		Success: success,
		History: p.History,
		State: string(p.State),
	}
}

func (p *PairOrchestrator) GetStatus() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return map[string]interface{}{
		"state": p.State,
		"task": p.Task,
		"history": p.History,
	}
}

func (p *PairOrchestrator) RotateRoles() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.Squad) < 3 { return }
	trio := p.Squad[:3]
	firstRole := trio[0].Role
	for i := 0; i < len(trio)-1; i++ { trio[i].Role = trio[i+1].Role }
	trio[len(trio)-1].Role = firstRole
}
