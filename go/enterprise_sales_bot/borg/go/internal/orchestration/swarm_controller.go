package orchestration

/**
 * @file swarm_controller.go
 * @module go/internal/orchestration
 *
 * WHAT: Go-native implementation of the Multi-Model Swarm Orchestration system.
 * Coordinates a team of LLMs on a shared goal with a shared transcript.
 *
 * WHY: Total Autonomy — The Go sidecar should be capable of managing 
 * complex multi-model workflows independently of the Node control plane.
 */

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/ai"
)

type SwarmRole string

const (
	SwarmRoleSupervisor  SwarmRole = "supervisor"
	SwarmRoleCritic      SwarmRole = "critic"
	SwarmRolePlanner     SwarmRole = "planner"
	SwarmRoleImplementer SwarmRole = "implementer"
	SwarmRoleTester      SwarmRole = "tester"
)

type SwarmMember struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Role     SwarmRole `json:"role"`
	Provider string    `json:"provider"`
	ModelID  string    `json:"modelId"`
	Status   string    `json:"status"` // "idle", "thinking", "working"
}

type SwarmSessionConfig struct {
	MaxTurns            int     `json:"maxTurns"`
	CompletionThreshold float64 `json:"completionThreshold"`
	AutoRotate          bool    `json:"autoRotate"`
}

type SwarmSessionResult struct {
	Success    bool     `json:"success"`
	Turns      int      `json:"turns"`
	Transcript []string `json:"transcript"`
}

type SwarmController struct {
	mu         sync.RWMutex
	members    map[string]*SwarmMember
	transcript []string
	activeGoal string
	broker     *A2ABroker
	bus        interface {
		EmitEvent(eventType string, source string, payload interface{})
	}
}

func NewSwarmController(broker *A2ABroker) *SwarmController {
	return &SwarmController{
		members:    make(map[string]*SwarmMember),
		transcript: make([]string, 0),
		broker:     broker,
	}
}

func (c *SwarmController) SetEventBus(bus interface {
	EmitEvent(eventType string, source string, payload interface{})
}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bus = bus
}

func (c *SwarmController) AddMember(member SwarmMember) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.members[member.ID] = &member
	fmt.Printf("[Go Swarm] Added member: %s (%s)\n", member.Name, member.Role)
}

func (c *SwarmController) StartSession(ctx context.Context, goal string, cfg SwarmSessionConfig) (*SwarmSessionResult, error) {
	c.mu.Lock()
	c.activeGoal = goal
	c.transcript = []string{fmt.Sprintf("Collective Goal: %s", goal)}
	c.mu.Unlock()

	fmt.Printf("[Go Swarm] Starting session for goal: %s\n", goal)

	turnCount := 0
	isComplete := false

	for turnCount < cfg.MaxTurns && !isComplete {
		turnCount++
		fmt.Printf("[Go Swarm] --- Turn %d ---\n", turnCount)

		// 1. Planning Turn
		plan, err := c.executeMemberTurn(ctx, SwarmRolePlanner, "Create the current implementation strategy.")
		if err != nil {
			return nil, err
		}
		c.addTranscript(fmt.Sprintf("PLANNER: %s", plan))

		// 2. Implementation Turn
		work, err := c.executeMemberTurn(ctx, SwarmRoleImplementer, fmt.Sprintf("Execute the plan: %s", plan))
		if err != nil {
			return nil, err
		}
		c.addTranscript(fmt.Sprintf("IMPLEMENTER: %s", work))

		// 3. Testing Turn
		testResult, err := c.executeMemberTurn(ctx, SwarmRoleTester, fmt.Sprintf("Verify the work: %s", work))
		if err != nil {
			return nil, err
		}
		c.addTranscript(fmt.Sprintf("TESTER: %s", testResult))

		// 4. Evaluation (Critic)
		evaluation, err := c.executeEvaluation(ctx, cfg.CompletionThreshold)
		if err != nil {
			return nil, err
		}
		c.addTranscript(fmt.Sprintf("CRITIC: %s", evaluation.Feedback))
		
		isComplete = evaluation.IsComplete
		
		if cfg.AutoRotate {
			c.rotateRoles()
		}

		c.broadcastUpdate()
	}

	return &SwarmSessionResult{
		Success:    isComplete,
		Turns:      turnCount,
		Transcript: c.GetTranscript(),
	}, nil
}

func (c *SwarmController) executeMemberTurn(ctx context.Context, role SwarmRole, instruction string) (string, error) {
	c.mu.RLock()
	var member *SwarmMember
	for _, m := range c.members {
		if m.Role == role {
			member = m
			break
		}
	}
	transcript := c.transcript
	bus := c.bus
	c.mu.RUnlock()

	if member == nil {
		return fmt.Sprintf("[System]: No active member for role %s", role), nil
	}

	if bus != nil {
		bus.EmitEvent("swarm:turn_start", "SwarmController", map[string]interface{}{
			"role":        string(role),
			"name":        member.Name,
			"modelId":     member.ModelID,
			"instruction": instruction,
		})
	}

	member.Status = "thinking"

	systemPrompt := ai.GetSwarmPrompt(string(role))

	prompt := fmt.Sprintf(`
		TRANSCRIPT:
		%s

		INSTRUCTION for %s (%s):
		%s
	`, strings.Join(transcript, "\n\n"), member.Name, strings.ToUpper(string(member.Role)), instruction)

	messages := []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}

	resp, err := ai.AutoRouteWithModel(ctx, member.ModelID, messages)
	member.Status = "idle"
	if err != nil {
		if bus != nil {
			bus.EmitEvent("swarm:turn_end", "SwarmController", map[string]interface{}{
				"role":    string(role),
				"success": false,
				"error":   err.Error(),
			})
		}
		return "", err
	}

	if bus != nil {
		bus.EmitEvent("swarm:turn_end", "SwarmController", map[string]interface{}{
			"role":    string(role),
			"success": true,
			"content": resp.Content,
		})
	}

	return resp.Content, nil
}

type SwarmEvaluation struct {
	IsComplete bool
	Feedback   string
}

func (c *SwarmController) executeEvaluation(ctx context.Context, threshold float64) (*SwarmEvaluation, error) {
	c.mu.RLock()
	var critic *SwarmMember
	for _, m := range c.members {
		if m.Role == SwarmRoleCritic {
			critic = m
			break
		}
	}
	transcript := c.transcript
	goal := c.activeGoal
	c.mu.RUnlock()

	modelID := "gemini-2.5-flash"
	if critic != nil {
		modelID = critic.ModelID
	}

	prompt := fmt.Sprintf(`
		Evaluate the following swarm transcript against the goal: "%s"
		
		TRANSCRIPT:
		%s

		Is the task complete? If so, start your response with "COMPLETE".
		Otherwise, provide constructive criticism for the next cycle.
	`, goal, strings.Join(transcript[len(transcript)-min(5, len(transcript)):], "\n\n"))

	messages := []ai.Message{
		{Role: "system", Content: ai.GetSwarmPrompt("critic")},
		{Role: "user", Content: prompt},
	}

	resp, err := ai.AutoRouteWithModel(ctx, modelID, messages)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(resp.Content)
	return &SwarmEvaluation{
		IsComplete: strings.HasPrefix(content, "COMPLETE"),
		Feedback:   content,
	}, nil
}

func (c *SwarmController) rotateRoles() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	ids := make([]string, 0, len(c.members))
	for id := range c.members {
		ids = append(ids, id)
	}
	if len(ids) < 2 {
		return
	}

	firstRole := c.members[ids[0]].Role
	for i := 0; i < len(ids)-1; i++ {
		c.members[ids[i]].Role = c.members[ids[i+1]].Role
	}
	c.members[ids[len(ids)-1]].Role = firstRole
}

func (c *SwarmController) addTranscript(entry string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.transcript = append(c.transcript, entry)
}

func (c *SwarmController) GetTranscript() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]string(nil), c.transcript...)
}

func (c *SwarmController) broadcastUpdate() {
	c.mu.RLock()
	members := make([]SwarmMember, 0, len(c.members))
	for _, m := range c.members {
		members = append(members, *m)
	}
	transcriptCount := len(c.transcript)
	activeGoal := c.activeGoal
	c.mu.RUnlock()

	c.broker.RouteMessage(A2AMessage{
		ID:        fmt.Sprintf("swarm-update-%d", nowMillis()),
		Timestamp: nowMillis(),
		Sender:    "SWARM_CONTROLLER",
		Type:      StateUpdate,
		Payload: map[string]interface{}{
			"members":         members,
			"transcriptCount": transcriptCount,
			"activeGoal":      activeGoal,
		},
	})
}


func nowMillis() int64 {
	return time.Now().UTC().UnixMilli()
}
