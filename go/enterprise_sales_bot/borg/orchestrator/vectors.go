package orchestrator

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/NexusSoftMDMA/TormentNexus/agents"
)

// VectorDB handles the "Jules Autopilot" requirement directly over SQLite natively
type VectorDB struct {
	db *sql.DB
}

func NewVectorDB(dbPath string) (*VectorDB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed binding sqlite for RAG: %w", err)
	}

	// Storing numerical slices as raw JSON text to avoid CGO-SQLite-vec compilation flaws on Windows.
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tool_embeddings (
		name TEXT PRIMARY KEY,
		description TEXT,
		embedding TEXT
	)`)
	if err != nil {
		return nil, fmt.Errorf("vector schema migration failed: %w", err)
	}

	return &VectorDB{db: db}, nil
}

// StoreTool calculates an embedding via external API (stubbed here) and saves it natively.
func (v *VectorDB) StoreTool(tool agents.Tool, vector []float64) error {
	vecBytes, _ := json.Marshal(vector)
	_, err := v.db.Exec("INSERT OR REPLACE INTO tool_embeddings (name, description, embedding) VALUES (?, ?, ?)",
		tool.Name, tool.Description, string(vecBytes))
	return err
}

// cosineSimilarity calculates purely mathematical alignment across native Go slices
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0.0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

type QueryResult struct {
	ToolName    string
	Description string
	Score       float64
}

// Search retrieves tools dynamically bypassing Context Window exhaustion (Progressive Disclosure Parity)
func (v *VectorDB) Search(intentVector []float64, topK int) ([]QueryResult, error) {
	rows, err := v.db.Query("SELECT name, description, embedding FROM tool_embeddings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []QueryResult

	for rows.Next() {
		var name, desc, embedStr string
		if err := rows.Scan(&name, &desc, &embedStr); err != nil {
			continue
		}

		var dbVec []float64
		if err := json.Unmarshal([]byte(embedStr), &dbVec); err != nil {
			continue // skip corrupted numerical arrays
		}

		score := cosineSimilarity(intentVector, dbVec)
		results = append(results, QueryResult{ToolName: name, Description: desc, Score: score})
	}

	// Fast memory pointer sort natively prioritizing highest relevance
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}
