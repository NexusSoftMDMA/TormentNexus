package graph

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
)

type NodeType string

const (
	NodeConcept  NodeType = "concept"
	NodeTool     NodeType = "tool"
	NodeProject  NodeType = "project"
	NodeDecision NodeType = "decision"
	NodePerson   NodeType = "person"
)

type EdgeType string

const (
	EdgeUses      EdgeType = "uses"
	EdgeDependsOn EdgeType = "depends_on"
	EdgeFixedBy   EdgeType = "fixed_by"
	EdgeRefersTo  EdgeType = "refers_to"
	EdgeMemberOf  EdgeType = "member_of"
)

type Node struct {
	ID        string                 `json:"id"`
	Type      NodeType               `json:"type"`
	Label     string                 `json:"label"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"createdAt"`
}

type Edge struct {
	FromID    string                 `json:"fromId"`
	ToID      string                 `json:"toId"`
	Type      EdgeType               `json:"type"`
	Weight    float64                `json:"weight"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"createdAt"`
}

type SemanticGraph struct {
	mu    sync.RWMutex
	nodes map[string]*Node
	edges []Edge
	vault controlplane.MemoryVault
}

func NewSemanticGraph(vault controlplane.MemoryVault) *SemanticGraph {
	return &SemanticGraph{
		nodes: make(map[string]*Node),
		edges: []Edge{},
		vault: vault,
	}
}

func (g *SemanticGraph) AddNode(n Node) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}
	g.nodes[n.ID] = &n
}

func (g *SemanticGraph) AddEdge(e Edge) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	g.edges = append(g.edges, e)
}

func (g *SemanticGraph) GetNeighbors(nodeID string) []Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var neighbors []Node
	for _, edge := range g.edges {
		if edge.FromID == nodeID {
			if n, ok := g.nodes[edge.ToID]; ok {
				neighbors = append(neighbors, *n)
			}
		}
	}
	return neighbors
}

// IngestFact extracts semantic nodes and edges from a raw memory content
func (g *SemanticGraph) IngestFact(ctx context.Context, content string) error {
	// This will eventually call an LLM to perform entity extraction.
	// For the scaffold, we just log the attempt.
	fmt.Printf("[Graph] Ingesting semantic fact: %s\n", content)
	return nil
}

type GraphStats struct {
	NodeCount int `json:"nodeCount"`
	EdgeCount int `json:"edgeCount"`
}

func (g *SemanticGraph) Stats() GraphStats {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return GraphStats{
		NodeCount: len(g.nodes),
		EdgeCount: len(g.edges),
	}
}
