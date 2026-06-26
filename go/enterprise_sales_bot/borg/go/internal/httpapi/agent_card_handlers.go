package httpapi

import (
    "net/http"
)

// AgentSkill represents a skill exposed via the A2A agent card.
type AgentSkill struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Tags        []string `json:"tags,omitempty"`
}

// handleAgentCard returns a JSON array of AgentSkill objects for
// FreeLLM A2A discovery. It pulls data from the in‑memory skillRegistry.
func (s *Server) handleAgentCard(w http.ResponseWriter, r *http.Request) {
    skills := s.skillRegistry.List()
    agentSkills := make([]AgentSkill, len(skills))
    for i, sk := range skills {
        agentSkills[i] = AgentSkill{
            ID:          sk.ID,
            Name:        sk.Name,
            Description: sk.Description,
            Tags:        []string{}, // Tags are not currently stored; left empty for now
        }
    }
    writeJSON(w, http.StatusOK, agentSkills)
}
