package orchestration

/**
 * @file pair_orchestrator.go
 * @module go/internal/orchestration
 *
 * WHAT: Multi-model pair programming state machine.
 * Strictly enforces the Planner -> Implementer -> Tester -> Critic turn cycle.
 *
 * WHY: Structured state transitions ensure that plans are vetted, implementations
 * are verified, and the final result is audited by a non-participant (Critic).
 */

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/borghq/borg-go/internal/ai"
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
}

func NewPairOrchestrator() *PairOrchestrator {
	return &PairOrchestrator{
		History: []string{},
		State:   StatePlanning,
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

func (p *PairOrchestrator) RotateRoles() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.Squad) < 3 {
		return
	}

	// Rotate main trio, keep Critic stable or rotate separately
	// For now, rotate Planner/Implementer/Tester
	trio := p.Squad[:3]
	firstRole := trio[0].Role
	for i := 0; i < len(trio)-1; i++ {
		trio[i].Role = trio[i+1].Role
	}
	trio[len(trio)-1].Role = firstRole

	fmt.Println("[PairOrchestrator] 🔄 Roles rotated:")
	for _, m := range p.Squad {
		fmt.Printf("  - %s: %s\n", m.Name, m.Role)
	}
}

func (p *PairOrchestrator) RunTask(ctx context.Context, task string) (*PairSessionResult, error) {
	p.mu.Lock()
	p.Task = task
	p.History = []string{"USER: " + task}
	p.State = StatePlanning
	p.mu.Unlock()

	fmt.Printf("[PairOrchestrator] 🚀 Starting Multi-Model Session: \"%s\"\n", task)

	for {
		p.mu.RLock()
		state := p.State
		p.mu.RUnlock()

		switch state {
		case StatePlanning:
			plan, err := p.executeTurn(ctx, Planner, "Create a detailed implementation plan for this task: "+p.Task)
			if err != nil {
				return p.failSession(err)
			}
			p.addHistory(fmt.Sprintf("PLANNER (%s): %s", p.getMemberName(Planner), plan))
			p.transition(StateReviewing)

		case StateReviewing:
			lastEntry := p.getLastHistory()
			feedback, err := p.executeTurn(ctx, Tester, "Review this plan and identify potential edge cases or bugs: "+lastEntry)
			if err != nil {
				return p.failSession(err)
			}
			p.addHistory(fmt.Sprintf("TESTER (%s): %s", p.getMemberName(Tester), feedback))
			p.transition(StateRefining)

		case StateRefining:
			lastEntry := p.getLastHistory()
			finalPlan, err := p.executeTurn(ctx, Planner, "Refine the plan based on this feedback: "+lastEntry)
			if err != nil {
				return p.failSession(err)
			}
			p.addHistory(fmt.Sprintf("PLANNER (%s): %s", p.getMemberName(Planner), finalPlan))
			p.transition(StateImplementing)

		case StateImplementing:
			lastEntry := p.getLastHistory()
			implementation, err := p.executeTurn(ctx, Implementer, "Implement the final plan. Focus on correctness and performance. Plan: "+lastEntry)
			if err != nil {
				return p.failSession(err)
			}
			p.addHistory(fmt.Sprintf("IMPLEMENTER (%s): %s", p.getMemberName(Implementer), implementation))
			p.transition(StateVerifying)

		case StateVerifying:
			lastEntry := p.getLastHistory()
			verification, err := p.executeTurn(ctx, Tester, "Verify the implementation against the plan and task requirements. Implementation: "+lastEntry)
			if err != nil {
				return p.failSession(err)
			}
			p.addHistory(fmt.Sprintf("TESTER (%s): %s", p.getMemberName(Tester), verification))
			p.transition(StateEvaluating)

		case StateEvaluating:
			// Final audit by the Critic
			audit, err := p.executeTurn(ctx, Critic, "Perform a final audit of the conversation and implementation. Does it satisfy the original goal? Respond with 'COMPLETE' if successful, or list remaining issues.")
			if err != nil {
				return p.failSession(err)
			}
			p.addHistory(fmt.Sprintf("CRITIC (%s): %s", p.getMemberName(Critic), audit))

			if strings.Contains(strings.ToUpper(audit), "COMPLETE") {
				p.transition(StateCompleted)
				return p.getResult(true), nil
			} else {
				p.transition(StateFailed)
				return p.getResult(false), nil
			}

		case StateCompleted:
			return p.getResult(true), nil

		case StateFailed:
			return p.getResult(false), nil

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

	if member == nil {
		return "", fmt.Errorf("no member assigned to role: %s", role)
	}

	if bus != nil {
		bus.EmitEvent("swarm:turn_start", "PairOrchestrator", map[string]interface{}{
			"role":   string(role),
			"name":   member.Name,
			"model":  member.ModelID,
			"prompt": prompt,
		})
	}

	fmt.Printf("[PairOrchestrator] 👤 %s (%s) is thinking...\n", member.Name, member.Role)

	systemPrompt := fmt.Sprintf(`You are part of a multi-agent pair programming squad. 
Your name is %s. Your current role is %s.
Collaborate with your teammates to solve the task perfectly.

SQUAD ROLES:
- PLANNER: Breaks down the task and designs the solution.
- IMPLEMENTER: Writes the actual code and executes tools.
- TESTER: Identifies bugs, edge cases, and verifies correctness.
- CRITIC: Audits the final output for absolute perfection.`, member.Name, strings.ToUpper(string(member.Role)))

	p.mu.RLock()
	fullHistory := strings.Join(p.History, "\n\n")
	p.mu.RUnlock()

	turnPrompt := fmt.Sprintf("CONVERSATION HISTORY:\n%s\n\nCURRENT TURN (%s): %s", fullHistory, strings.ToUpper(string(member.Role)), prompt)

	resp, err := ai.AutoRouteWithModel(ctx, member.Provider+"/"+member.ModelID, []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: turnPrompt},
	})

	if err != nil {
		fmt.Printf("[PairOrchestrator] ⚠️ Turn failed for %s: %v\n", member.Name, err)
		if bus != nil {
			bus.EmitEvent("swarm:turn_end", "PairOrchestrator", map[string]interface{}{
				"role":    string(role),
				"success": false,
				"error":   err.Error(),
			})
		}
		return "", err
	}

	if bus != nil {
		bus.EmitEvent("swarm:turn_end", "PairOrchestrator", map[string]interface{}{
			"role":    string(role),
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
	if len(p.History) == 0 {
		return ""
	}
	return p.History[len(p.History)-1]
}

func (p *PairOrchestrator) transition(newState SessionState) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Printf("[PairOrchestrator] ⚙️ State transition: %s -> %s\n", p.State, newState)
	p.State = newState
}

func (p *PairOrchestrator) getMemberName(role PairRole) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, m := range p.Squad {
		if m.Role == role {
			return m.Name
		}
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
	
	finalOutput := ""
	for i := len(p.History) - 1; i >= 0; i-- {
		if strings.HasPrefix(p.History[i], "IMPLEMENTER") {
			finalOutput = strings.TrimPrefix(p.History[i], "IMPLEMENTER")
			if idx := strings.Index(finalOutput, ": "); idx != -1 {
				finalOutput = finalOutput[idx+2:]
			}
			break
		}
	}

	return &PairSessionResult{
		Success:     success,
		History:     p.History,
		FinalOutput: finalOutput,
		State:       string(p.State),
	}
}

// GetStatus returns a thread-safe snapshot of the orchestrator's current state.
func (p *PairOrchestrator) GetStatus() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return map[string]interface{}{
		"state":   p.State,
		"task":    p.Task,
		"history": p.History,
		"squad":   p.Squad,
	}
}
