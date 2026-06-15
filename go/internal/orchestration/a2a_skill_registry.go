package orchestration

/**
 * @file a2a_skill_registry.go
 * @module go/internal/orchestration
 *
 * WHAT: A2A Skill Registry that maps skill IDs to agent URLs for the FreeLLM swarm.
 * Stores which agent (URL) provides which skill, and enables findAgentForSkill lookups.
 *
 * WHY: The FreeLLM A2A swarm uses AgentSkill structs (ID, Name, Description, Tags)
 * declared in agent cards. The SwarmCoordinator queries findAgentForSkill(skillID)
 * to dispatch tasks to the right agent.
 */

import (
	"sync"
	"time"
)

// A2ASkillRegistry maintains the mapping of skill IDs to agents that provide them.
// It mirrors the FreeLLM SwarmCoordinator's in-memory registry but is Go-native.
type A2ASkillRegistry struct {
	mu sync.RWMutex

	// skillAgents maps skill ID → list of agent URLs that provide it (newest first)
	skillAgents map[string][]string

	// agentSkills maps agent URL → skill IDs it provides
	agentSkills map[string][]string

	// agentLastSeen tracks the last-seen timestamp per agent URL
	agentLastSeen map[string]int64
}

// NewA2ASkillRegistry creates a new empty A2A skill registry.
func NewA2ASkillRegistry() *A2ASkillRegistry {
	return &A2ASkillRegistry{
		skillAgents:  make(map[string][]string),
		agentSkills:  make(map[string][]string),
		agentLastSeen: make(map[string]int64),
	}
}

// RegisterAgentSkill registers a skill ID as being provided by the given agent URL.
// If the agent is already in the list, it is moved to the front (most recent).
func (r *A2ASkillRegistry) RegisterAgentSkill(agentURL, skillID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Add to skillAgents: prepend if not already present
	list := r.skillAgents[skillID]
	for i, url := range list {
		if url == agentURL {
			// Move to front
			r.skillAgents[skillID] = append([]string{agentURL}, append(list[:i], list[i+1:]...)...)
			r.agentLastSeen[agentURL] = time.Now().UnixMilli()
			return
		}
	}
	r.skillAgents[skillID] = append([]string{agentURL}, list...)

	// Add to agentSkills
	skills := r.agentSkills[agentURL]
	for _, s := range skills {
		if s == skillID {
			return
		}
	}
	r.agentSkills[agentURL] = append(skills, skillID)
	r.agentLastSeen[agentURL] = time.Now().UnixMilli()
}

// UnregisterAgent removes an agent and all its skills from the registry.
func (r *A2ASkillRegistry) UnregisterAgent(agentURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove from agentSkills
	if skills, ok := r.agentSkills[agentURL]; ok {
		for _, skillID := range skills {
			list := r.skillAgents[skillID]
			for i, url := range list {
				if url == agentURL {
					// Remove from list
					r.skillAgents[skillID] = append(list[:i], list[i+1:]...)
					break
				}
			}
		}
	}
	delete(r.agentSkills, agentURL)
	delete(r.agentLastSeen, agentURL)
}

// findAgentForSkill returns the most recently seen agent URL that provides the given skill.
// Returns empty string if no agent provides this skill.
func (r *A2ASkillRegistry) FindAgentForSkill(skillID string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list, ok := r.skillAgents[skillID]
	if !ok || len(list) == 0 {
		return ""
	}
	return list[0] // most recent
}

// ListSkillAgents returns all agent URLs that provide the given skill.
func (r *A2ASkillRegistry) ListSkillAgents(skillID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list, ok := r.skillAgents[skillID]
	if !ok {
		return []string{}
	}
	result := make([]string, len(list))
	copy(result, list)
	return result
}

// ListAgentSkills returns all skill IDs provided by the given agent.
func (r *A2ASkillRegistry) ListAgentSkills(agentURL string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills, ok := r.agentSkills[agentURL]
	if !ok {
		return []string{}
	}
	result := make([]string, len(skills))
	copy(result, skills)
	return result
}

// ListAllSkillAgents returns a copy of the entire skill→agents mapping.
func (r *A2ASkillRegistry) ListAllSkillAgents() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]string, len(r.skillAgents))
	for k, v := range r.skillAgents {
		result[k] = v
	}
	return result
}

// AgentCount returns the number of registered agents.
func (r *A2ASkillRegistry) AgentCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agentSkills)
}

// SkillCount returns the number of registered skills.
func (r *A2ASkillRegistry) SkillCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.skillAgents)
}