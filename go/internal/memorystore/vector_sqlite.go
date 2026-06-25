package memorystore

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
	_ "modernc.org/sqlite"
)

type l1Entry struct {
	value      controlplane.L2VaultRecord
	heat       float64
	lastAccess time.Time
}

type QueryPayload struct {
	QueryText string    `json:"query_text"`
	QueryVec  []float32 `json:"query_vec"`
	Kind      string    `json:"kind"`
	Category  string    `json:"category"`
	Tags      []string  `json:"tags"`
}

type VectorStore struct {
	db      *sql.DB
	mu      sync.Mutex
	l1Cache map[string]*l1Entry
	l1Max   int
}

func NewVectorStore(dbPath string) (*VectorStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if dbPath != ":memory:" {
		if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set WAL mode: %w", err)
		}
		if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set synchronous mode: %w", err)
		}
		if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set busy timeout: %w", err)
		}
	}

	if _, err := db.Exec(controlplane.VectorSchemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init vector schema: %w", err)
	}

	return &VectorStore{
		db:      db,
		l1Cache: make(map[string]*l1Entry),
		l1Max:   100,
	}, nil
}

func (s *VectorStore) Close() error {
	return s.db.Close()
}

func (s *VectorStore) Commit(ctx context.Context, entry controlplane.L2VaultRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry.HeatScore == 0 {
		entry.HeatScore = 50.0
	}
	if entry.LastAccessedAt.IsZero() {
		entry.LastAccessedAt = time.Now()
	}
	if entry.Kind == "" {
		entry.Kind = "fact"
	}
	if entry.Category == "" {
		entry.Category = "general"
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO l2_vault (id, session_id, memory_type, memory_kind, category, tags, source_url, content, importance, heat_score, last_accessed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			content = excluded.content,
			memory_kind = excluded.memory_kind,
			category = excluded.category,
			tags = excluded.tags,
			source_url = excluded.source_url,
			importance = excluded.importance,
			heat_score = excluded.heat_score,
			last_accessed_at = excluded.last_accessed_at,
			created_at = excluded.created_at
	`, entry.ID, entry.SessionID, string(entry.Type), entry.Kind, entry.Category, entry.Tags, entry.SourceURL, entry.Content, entry.Importance, entry.HeatScore, entry.LastAccessedAt, entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("memorystore commit insert: %w", err)
	}

	// Update L1 cache
	if len(s.l1Cache) >= s.l1Max {
		s.evictColdestL1Locked()
	}
	s.l1Cache[entry.ID] = &l1Entry{
		value:      entry,
		heat:       1.0,
		lastAccess: time.Now(),
	}

	if len(entry.Embedding) > 0 {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO vec_l2_vault (id, embedding)
			VALUES (?, ?)
			ON CONFLICT(id) DO UPDATE SET embedding = excluded.embedding
		`, entry.ID, encodeVec(entry.Embedding))
		if err != nil {
			return fmt.Errorf("memorystore commit embedding: %w", err)
		}
	}
	return nil
}

func (s *VectorStore) evictColdestL1Locked() {
	if len(s.l1Cache) == 0 {
		return
	}
	var coldestKey string
	minHeat := math.MaxFloat64
	for k, e := range s.l1Cache {
		if e.heat < minHeat {
			minHeat = e.heat
			coldestKey = k
		}
	}
	delete(s.l1Cache, coldestKey)
}

func (s *VectorStore) SemanticSearch(ctx context.Context, query string, limit int) ([]controlplane.L2VaultRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Try parsing structured query payload or direct JSON float array
	var queryPayload QueryPayload
	isStructured := false
	trimmedQuery := strings.TrimSpace(query)

	if strings.HasPrefix(trimmedQuery, "{") {
		if err := json.Unmarshal([]byte(query), &queryPayload); err == nil {
			isStructured = true
		}
	}

	var queryVec []float32
	var queryText string
	var filterKind string
	var filterCategory string

	if isStructured {
		queryVec = queryPayload.QueryVec
		queryText = queryPayload.QueryText
		filterKind = queryPayload.Kind
		filterCategory = queryPayload.Category
	} else if strings.HasPrefix(trimmedQuery, "[") {
		if err := json.Unmarshal([]byte(query), &queryVec); err == nil && len(queryVec) > 0 {
			// successfully parsed vector directly
		}
	} else {
		queryText = query
	}

	isVectorSearch := len(queryVec) > 0

	if isVectorSearch {
		// Vector search with optional metadata filters
		var args []interface{}
		sqlQuery := `
			SELECT v.id, v.embedding, l.session_id, l.memory_type, l.memory_kind, l.category, l.tags, l.source_url, l.content, l.importance, l.heat_score, l.last_accessed_at, l.created_at
			FROM vec_l2_vault v
			JOIN l2_vault l ON l.id = v.id
			WHERE l.memory_type != 'archive'
		`
		if filterKind != "" {
			sqlQuery += " AND l.memory_kind = ?"
			args = append(args, filterKind)
		}
		if filterCategory != "" {
			sqlQuery += " AND l.category = ?"
			args = append(args, filterCategory)
		}

		rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
		if err != nil {
			return nil, fmt.Errorf("memorystore vector search: %w", err)
		}
		defer rows.Close()

		type scored struct {
			record controlplane.L2VaultRecord
			score  float64
		}
		var candidates []scored

		for rows.Next() {
			var r controlplane.L2VaultRecord
			var blob []byte
			var mType string
			if err := rows.Scan(&r.ID, &blob, &r.SessionID, &mType, &r.Kind, &r.Category, &r.Tags, &r.SourceURL, &r.Content, &r.Importance, &r.HeatScore, &r.LastAccessedAt, &r.CreatedAt); err != nil {
				return nil, err
			}
			r.Type = controlplane.MemoryType(mType)
			
			vec := decodeVec(blob, len(blob)/4)
			sim := cosineSim(queryVec, vec)
			
			// Boost score slightly using importance
			boostedSim := sim * (0.8 + 0.2*r.Importance)
			if boostedSim >= 0.3 {
				candidates = append(candidates, scored{record: r, score: boostedSim})
			}
		}

		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].score > candidates[j].score
		})

		if len(candidates) > limit {
			candidates = candidates[:limit]
		}

		results := make([]controlplane.L2VaultRecord, len(candidates))
		for i, c := range candidates {
			results[i] = c.record
			s.incrementHeatLocked(ctx, c.record.ID)
		}
		return results, nil
	}

	// Check L1 cache first for manual / working memory queries (supporting text filter)
	if queryText != "" {
		var l1Results []controlplane.L2VaultRecord
		for _, e := range s.l1Cache {
			match := strings.Contains(strings.ToLower(e.value.Content), strings.ToLower(queryText))
			if filterKind != "" && e.value.Kind != filterKind {
				match = false
			}
			if filterCategory != "" && e.value.Category != filterCategory {
				match = false
			}
			if match && e.value.Type != controlplane.MemoryArchive {
				e.heat += 1.0
				e.lastAccess = time.Now()
				l1Results = append(l1Results, e.value)
			}
		}
		if len(l1Results) > 0 {
			sort.Slice(l1Results, func(i, j int) bool {
				return l1Results[i].Importance > l1Results[j].Importance
			})
			if len(l1Results) > limit {
				l1Results = l1Results[:limit]
			}
			return l1Results, nil
		}
	}

	// Fall back to keyword search with optional filters
	var args []interface{}
	sqlQuery := `
		SELECT id, session_id, memory_type, memory_kind, category, tags, source_url, content, importance, heat_score, last_accessed_at, created_at
		FROM l2_vault
		WHERE memory_type != 'archive'
	`
	if queryText != "" {
		sqlQuery += " AND content LIKE ?"
		args = append(args, "%"+queryText+"%")
	}
	if filterKind != "" {
		sqlQuery += " AND memory_kind = ?"
		args = append(args, filterKind)
	}
	if filterCategory != "" {
		sqlQuery += " AND category = ?"
		args = append(args, filterCategory)
	}
	sqlQuery += " ORDER BY importance DESC, heat_score DESC, created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("memorystore search: %w", err)
	}
	defer rows.Close()

	var results []controlplane.L2VaultRecord
	for rows.Next() {
		var r controlplane.L2VaultRecord
		var mType string
		if err := rows.Scan(&r.ID, &r.SessionID, &mType, &r.Kind, &r.Category, &r.Tags, &r.SourceURL, &r.Content, &r.Importance, &r.HeatScore, &r.LastAccessedAt, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.Type = controlplane.MemoryType(mType)
		results = append(results, r)
	}

	// Update heat and last_accessed_at for hits
	for _, r := range results {
		s.incrementHeatLocked(ctx, r.ID)
	}

	return results, nil
}

func (s *VectorStore) ReinforceMemory(ctx context.Context, id string, success bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var heatScore, importance float64
	err := s.db.QueryRowContext(ctx, "SELECT heat_score, importance FROM l2_vault WHERE id = ?", id).Scan(&heatScore, &importance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	if success {
		heatScore = math.Min(100.0, heatScore+15.0)
		importance = math.Min(1.0, importance+0.1)
	} else {
		heatScore = math.Max(0.0, heatScore-20.0)
		importance = math.Max(0.0, importance-0.2)
	}

	_, err = s.db.ExecContext(ctx, "UPDATE l2_vault SET heat_score = ?, importance = ?, last_accessed_at = CURRENT_TIMESTAMP WHERE id = ?", heatScore, importance, id)
	return err
}

func (s *VectorStore) GetVaultRecordCount(ctx context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM l2_vault").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("GetVaultRecordCount: %w", err)
	}
	return count, nil
}

func (s *VectorStore) incrementHeatLocked(ctx context.Context, id string) {
	_, _ = s.db.ExecContext(ctx, `
		UPDATE l2_vault
		SET heat_score = MIN(100.0, heat_score + 10.0),
		    last_accessed_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, id)
}

func (s *VectorStore) ApplyDecay(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// heat_score = heat_score * exp(-0.0288 * hours_since_access)
	_, err := s.db.ExecContext(ctx, `
		UPDATE l2_vault
		SET heat_score = heat_score * exp(-0.0288 * (julianday('now') - julianday(last_accessed_at)) * 24.0)
		WHERE memory_type != 'archive'
	`)
	if err != nil {
		return fmt.Errorf("apply decay: %w", err)
	}

	// Promote: Working memories with a heat > 80 move to long_term
	_, err = s.db.ExecContext(ctx, `
		UPDATE l2_vault
		SET memory_type = 'long_term'
		WHERE heat_score > 80.0 AND memory_type = 'working'
	`)
	if err != nil {
		return fmt.Errorf("promotion: %w", err)
	}

	// Demote: long_term memories with a heat < 20 move to the archive (L3)
	_, err = s.db.ExecContext(ctx, `
		UPDATE l2_vault
		SET memory_type = 'archive'
		WHERE heat_score < 20.0 AND memory_type = 'long_term'
	`)

	return err
}

func (s *VectorStore) GetAllVaultRecords(ctx context.Context, limit int) ([]controlplane.L2VaultRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, session_id, memory_type, memory_kind, category, tags, source_url, content, importance, heat_score, last_accessed_at, created_at
		FROM l2_vault
		ORDER BY created_at DESC
		LIMIT ?
	`)
	if err != nil {
		return nil, fmt.Errorf("GetAllVaultRecords: %w", err)
	}
	defer rows.Close()

	var results []controlplane.L2VaultRecord
	for rows.Next() {
		var r controlplane.L2VaultRecord
		var mType string
		if err := rows.Scan(&r.ID, &r.SessionID, &mType, &r.Kind, &r.Category, &r.Tags, &r.SourceURL, &r.Content, &r.Importance, &r.HeatScore, &r.LastAccessedAt, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.Type = controlplane.MemoryType(mType)
		results = append(results, r)
	}

	// Update heat and last_accessed_at for hits
	for _, r := range results {
		s.incrementHeatLocked(ctx, r.ID)
	}

	return results, nil
}

// Helpers for Vector Encoding and Cosine Similarity

func encodeVec(v []float32) []byte {
	buf := make([]byte, len(v)*4)
	for i, f := range v {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
	}
	return buf
}

func decodeVec(buf []byte, dim int) []float32 {
	if len(buf) < dim*4 {
		dim = len(buf) / 4
	}
	v := make([]float32, dim)
	for i := 0; i < dim; i++ {
		v[i] = math.Float32frombits(binary.LittleEndian.Uint32(buf[i*4:]))
	}
	return v
}

func cosineSim(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, nA, nB float64
	for i := range a {
		af := float64(a[i])
		bf := float64(b[i])
		dot += af * bf
		nA += af * af
		nB += bf * bf
	}
	if nA == 0 || nB == 0 {
		return 0
	}
	return dot / (math.Sqrt(nA) * math.Sqrt(nB))
}

// AddRelation creates or updates a relational edge between two L2 memories (GraphRAG relation mapping).
func (s *VectorStore) AddRelation(ctx context.Context, sourceID, targetID, relType string, weight float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO l2_relations (source_id, target_id, relation_type, weight)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(source_id, target_id, relation_type) DO UPDATE SET weight = excluded.weight
	`, sourceID, targetID, relType, weight)
	return err
}

// GetRelations returns all incoming and outgoing relations for a specific memory ID.
func (s *VectorStore) GetRelations(ctx context.Context, id string) ([]controlplane.L2Relation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.QueryContext(ctx, `
		SELECT source_id, target_id, relation_type, weight
		FROM l2_relations
		WHERE source_id = ? OR target_id = ?
	`, id, id)
	if err != nil {
		return nil, fmt.Errorf("GetRelations: %w", err)
	}
	defer rows.Close()

	var rels []controlplane.L2Relation
	for rows.Next() {
		var r controlplane.L2Relation
		if err := rows.Scan(&r.SourceID, &r.TargetID, &r.RelationType, &r.Weight); err != nil {
			return nil, err
		}
		rels = append(rels, r)
	}
	return rels, nil
}

// ForgettingCurveDecay applies Ebbinghaus biomimetic decay to the heat score of all non-archived memories
// based on the hours elapsed since last_accessed_at: Heat = Heat * exp(-0.0288 * hours_since_access).
func (s *VectorStore) ForgettingCurveDecay(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `
		UPDATE l2_vault
		SET heat_score = heat_score * exp(-0.0288 * (julianday('now') - julianday(last_accessed_at)) * 24.0)
		WHERE memory_type != 'archive'
	`)
	if err != nil {
		return fmt.Errorf("ForgettingCurveDecay: %w", err)
	}

	// Archive demotion: long_term memories with a heat score < 20.0 move to archive
	_, err = s.db.ExecContext(ctx, `
		UPDATE l2_vault
		SET memory_type = 'archive'
		WHERE heat_score < 20.0 AND memory_type = 'long_term'
	`)
	if err != nil {
		return fmt.Errorf("ForgettingCurveDecay archive promotion: %w", err)
	}

	return nil
}

// ConsolidateMemories performs semantic consolidation. It finds pairs of memories with similar content (using Jaccard similarity threshold > 0.8) and merges them.
func (s *VectorStore) ConsolidateMemories(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, content, COALESCE(heat_score, 0.0), COALESCE(importance, 0.0), memory_kind, category
		FROM l2_vault
		WHERE memory_type != 'archive'
	`)
	if err != nil {
		return fmt.Errorf("ConsolidateMemories fetch: %w", err)
	}
	defer rows.Close()

	type rawRecord struct {
		ID         string
		Content    string
		HeatScore  float64
		Importance float64
		Kind       string
		Category   string
	}

	var recs []rawRecord
	for rows.Next() {
		var r rawRecord
		if err := rows.Scan(&r.ID, &r.Content, &r.HeatScore, &r.Importance, &r.Kind, &r.Category); err != nil {
			return err
		}
		recs = append(recs, r)
	}

	jaccardSim := func(s1, s2 string) float64 {
		w1 := strings.Fields(strings.ToLower(s1))
		w2 := strings.Fields(strings.ToLower(s2))
		set1 := make(map[string]bool)
		set2 := make(map[string]bool)
		for _, w := range w1 {
			set1[w] = true
		}
		for _, w := range w2 {
			set2[w] = true
		}
		intersect := 0
		for w := range set1 {
			if set2[w] {
				intersect++
			}
		}
		union := len(set1) + len(set2) - intersect
		if union == 0 {
			return 0
		}
		return float64(intersect) / float64(union)
	}

	merged := make(map[string]bool)

	for i := 0; i < len(recs); i++ {
		if merged[recs[i].ID] {
			continue
		}
		for j := i + 1; j < len(recs); j++ {
			if merged[recs[j].ID] {
				continue
			}

			sim := jaccardSim(recs[i].Content, recs[j].Content)
			if sim > 0.8 {
				// Merge j into i
				// Combine content, boost heat/importance
				newContent := recs[i].Content
				if len(recs[j].Content) > len(recs[i].Content) {
					newContent = recs[j].Content
				}
				newHeat := math.Min(100.0, recs[i].HeatScore+recs[j].HeatScore*0.5)
				newImportance := math.Min(1.0, recs[i].Importance+0.1)

				_, err = s.db.ExecContext(ctx, `
					UPDATE l2_vault
					SET content = ?, heat_score = ?, importance = ?, last_accessed_at = CURRENT_TIMESTAMP
					WHERE id = ?
				`, newContent, newHeat, newImportance, recs[i].ID)
				if err != nil {
					return fmt.Errorf("ConsolidateMemories update: %w", err)
				}

				// Copy relations of j to i, then delete j
				_, err = s.db.ExecContext(ctx, `
					UPDATE l2_relations
					SET source_id = ?
					WHERE source_id = ?
				`, recs[i].ID, recs[j].ID)
				if err != nil {
					return err
				}
				_, err = s.db.ExecContext(ctx, `
					UPDATE l2_relations
					SET target_id = ?
					WHERE target_id = ?
				`, recs[i].ID, recs[j].ID)
				if err != nil {
					return err
				}

				_, err = s.db.ExecContext(ctx, "DELETE FROM l2_vault WHERE id = ?", recs[j].ID)
				if err != nil {
					return err
				}
				_, err = s.db.ExecContext(ctx, "DELETE FROM vec_l2_vault WHERE id = ?", recs[j].ID)
				if err != nil {
					return err
				}

				merged[recs[j].ID] = true
				recs[i].Content = newContent
				recs[i].HeatScore = newHeat
				recs[i].Importance = newImportance
			}
		}
	}

	return nil
}

