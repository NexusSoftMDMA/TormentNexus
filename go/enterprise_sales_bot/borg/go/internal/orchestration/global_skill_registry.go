package orchestration

/**
 * @file global_skill_registry.go
 * @module go/internal/orchestration
 *
 * WHAT: Global singleton for the A2A Skill Registry.
 * Exposes FindAgentForSkill as a top-level function for swarm dispatch.
 *
 * WHY: The FreeLLM SwarmCoordinator calls findAgentForSkill(skillID) to
 * discover which agent URL provides a given skill during task dispatch.
 */

// GlobalSkillRegistry is the package-level singleton that maps skill IDs
// to agent URLs for the FreeLLM A2A swarm. All agents / components should
// register their skills into this registry on startup.
var GlobalSkillRegistry = NewA2ASkillRegistry()

// FindAgentForSkill returns the most recently seen agent URL that provides
// the given skill ID. It delegates to GlobalSkillRegistry.FindAgentForSkill.
// Returns an empty string if no agent is found.
func FindAgentForSkill(skillID string) string {
	return GlobalSkillRegistry.FindAgentForSkill(skillID)
}
