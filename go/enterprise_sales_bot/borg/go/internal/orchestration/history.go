package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type DebateRecord struct {
	ID        string         `json:"id"`
	Timestamp int64          `json:"timestamp"`
	Task      map[string]any `json:"task"`
	Decision  map[string]any `json:"decision"`
	Metadata  DebateMetadata `json:"metadata"`
}

type DebateMetadata struct {
	SessionID                string         `json:"sessionId,omitempty"`
	DebateRounds             int            `json:"debateRounds"`
	ConsensusMode            string         `json:"consensusMode"`
	LeadSupervisor           string         `json:"leadSupervisor,omitempty"`
	DynamicSelection         map[string]any `json:"dynamicSelection,omitempty"`
	DurationMs               int64          `json:"durationMs"`
	SupervisorCount          int            `json:"supervisorCount"`
	ParticipatingSupervisors []string       `json:"participatingSupervisors"`
}

type DebateHistoryConfig struct {
	Enabled       bool   `json:"enabled"`
	StorageDir    string `json:"storageDir"`
	MaxRecords    int    `json:"maxRecords"`
	AutoSave      bool   `json:"autoSave"`
	RetentionDays int    `json:"retentionDays"`
}

type DebateQueryOptions struct {
	SessionID      string
	TaskType       string
	Approved       *bool
	SupervisorName string
	FromTimestamp  *int64
	ToTimestamp    *int64
	MinConsensus   *float64
	MaxConsensus   *float64
	Limit          int
	Offset         int
	SortBy         string
	SortOrder      string
}

type DebateStats struct {
	TotalDebates           int            `json:"totalDebates"`
	ApprovedCount          int            `json:"approvedCount"`
	RejectedCount          int            `json:"rejectedCount"`
	ApprovalRate           float64        `json:"approvalRate"`
	AverageConsensus       float64        `json:"averageConsensus"`
	AverageDurationMs      float64        `json:"averageDurationMs"`
	DebatesByTaskType      map[string]int `json:"debatesByTaskType"`
	DebatesBySupervisor    map[string]int `json:"debatesBySupervisor"`
	DebatesByConsensusMode map[string]int `json:"debatesByConsensusMode"`
	OldestDebate           *int64         `json:"oldestDebate,omitempty"`
	NewestDebate           *int64         `json:"newestDebate,omitempty"`
}

type DebateHistoryStore struct {
	dbPath string
	mu     sync.RWMutex
	config DebateHistoryConfig
}

func NewDebateHistoryStore(dbPath string) *DebateHistoryStore {
	return &DebateHistoryStore{
		dbPath: dbPath,
		config: DebateHistoryConfig{
			Enabled:       true,
			StorageDir:    filepath.Dir(dbPath),
			MaxRecords:    1000,
			AutoSave:      true,
			RetentionDays: 90,
		},
	}
}

func (s *DebateHistoryStore) Initialize(ctx context.Context) (int, error) {
	return s.GetRecordCount(ctx)
}

func (s *DebateHistoryStore) GetConfig() DebateHistoryConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *DebateHistoryStore) UpdateConfig(patch map[string]any) DebateHistoryConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	if value, ok := patch["enabled"].(bool); ok {
		s.config.Enabled = value
	}
	if value, ok := intValue(patch["maxRecords"]); ok && value > 0 {
		s.config.MaxRecords = value
	}
	if value, ok := intValue(patch["retentionDays"]); ok && value > 0 {
		s.config.RetentionDays = value
	}
	if value, ok := patch["autoSave"].(bool); ok {
		s.config.AutoSave = value
	}
	if value, ok := patch["storageDir"].(string); ok && strings.TrimSpace(value) != "" {
		s.config.StorageDir = value
	}
	return s.config
}

func (s *DebateHistoryStore) Toggle(enabled *bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if enabled == nil {
		s.config.Enabled = !s.config.Enabled
	} else {
		s.config.Enabled = *enabled
	}
	return s.config.Enabled
}

func (s *DebateHistoryStore) SaveNativeDebate(ctx context.Context, sessionID string, objective string, contextText string, result *DebateResult) (*DebateRecord, error) {
	if result == nil {
		return nil, fmt.Errorf("missing debate result")
	}
	participants := make([]string, 0, len(result.Contributions))
	decisionVotes := make([]map[string]any, 0, len(result.Contributions))
	for _, contribution := range result.Contributions {
		participants = append(participants, contribution.Role)
		decisionVotes = append(decisionVotes, map[string]any{
			"supervisor": contribution.Role,
			"message":    contribution.Message,
		})
	}
	record := &DebateRecord{
		ID:        "debate_" + uuid.NewString(),
		Timestamp: time.Now().UTC().UnixMilli(),
		Task: map[string]any{
			"description": objective,
			"context":     contextText,
		},
		Decision: map[string]any{
			"approved":          result.Approved,
			"finalPlan":         result.FinalPlan,
			"consensus":         result.Consensus,
			"weightedConsensus": result.Consensus,
			"votes":             decisionVotes,
		},
		Metadata: DebateMetadata{
			SessionID:                sessionID,
			DebateRounds:             len(result.Contributions),
			ConsensusMode:            "weighted",
			LeadSupervisor:           firstString(participants),
			DurationMs:               0,
			SupervisorCount:          len(participants),
			ParticipatingSupervisors: participants,
		},
	}
	if !s.GetConfig().Enabled {
		return record, nil
	}
	if err := s.SaveRecord(ctx, *record); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *DebateHistoryStore) SaveRecord(ctx context.Context, record DebateRecord) error {
	db, err := s.open(ctx)
	if err != nil {
		return err
	}
	defer db.Close()
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}
	outcome := "rejected"
	if approved, _ := record.Decision["approved"].(bool); approved {
		outcome = "approved"
	}
	consensus := floatValueOr(record.Decision["consensus"], 0)
	weightedConsensus := floatValueOr(record.Decision["weightedConsensus"], consensus)
	taskType := stringValue(record.Metadata.DynamicSelection["taskType"])
	if taskType == "" {
		taskType = "general"
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO council_debates (id, title, session_id, workspace_id, task_type, status, consensus, weighted_consensus, outcome, rounds, timestamp, data)
		VALUES (?, ?, ?, NULL, ?, 'completed', ?, ?, ?, ?, ?, ?)
	`, record.ID, truncateString(stringValue(record.Task["description"]), 255), nullIfEmpty(record.Metadata.SessionID), taskType, consensus, weightedConsensus, outcome, record.Metadata.DebateRounds, record.Timestamp, string(payload))
	if err != nil {
		return err
	}
	return s.prune(ctx, db)
}

func (s *DebateHistoryStore) GetRecordCount(ctx context.Context) (int, error) {
	db, err := s.open(ctx)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM council_debates`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *DebateHistoryStore) GetStorageSize() int64 {
	info, err := os.Stat(s.dbPath)
	if err != nil {
		return 0
	}
	return info.Size()
}

func (s *DebateHistoryStore) GetDebate(ctx context.Context, id string) (*DebateRecord, error) {
	db, err := s.open(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return querySingleDebate(ctx, db, `SELECT data FROM council_debates WHERE id = ? LIMIT 1`, id)
}

func (s *DebateHistoryStore) QueryDebates(ctx context.Context, options DebateQueryOptions) ([]DebateRecord, int, error) {
	db, err := s.open(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer db.Close()
	whereClauses := []string{"1=1"}
	args := make([]any, 0)
	if strings.TrimSpace(options.SessionID) != "" {
		whereClauses = append(whereClauses, "session_id = ?")
		args = append(args, options.SessionID)
	}
	if strings.TrimSpace(options.TaskType) != "" {
		whereClauses = append(whereClauses, "task_type = ?")
		args = append(args, options.TaskType)
	}
	if options.Approved != nil {
		if *options.Approved {
			whereClauses = append(whereClauses, "outcome = 'approved'")
		} else {
			whereClauses = append(whereClauses, "outcome = 'rejected'")
		}
	}
	if options.FromTimestamp != nil {
		whereClauses = append(whereClauses, "timestamp >= ?")
		args = append(args, *options.FromTimestamp)
	}
	if options.ToTimestamp != nil {
		whereClauses = append(whereClauses, "timestamp <= ?")
		args = append(args, *options.ToTimestamp)
	}
	if options.MinConsensus != nil {
		whereClauses = append(whereClauses, "consensus >= ?")
		args = append(args, *options.MinConsensus)
	}
	if options.MaxConsensus != nil {
		whereClauses = append(whereClauses, "consensus <= ?")
		args = append(args, *options.MaxConsensus)
	}
	query := `SELECT data FROM council_debates WHERE ` + strings.Join(whereClauses, " AND ")
	sortBy := "timestamp"
	switch options.SortBy {
	case "consensus":
		sortBy = "consensus"
	case "duration":
		sortBy = "rounds"
	}
	sortOrder := "DESC"
	if strings.EqualFold(options.SortOrder, "asc") {
		sortOrder = "ASC"
	}
	query += ` ORDER BY ` + sortBy + ` ` + sortOrder
	limit := options.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := options.Offset
	query += ` LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	records := make([]DebateRecord, 0)
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, 0, err
		}
		var record DebateRecord
		if err := json.Unmarshal([]byte(raw), &record); err != nil {
			continue
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if strings.TrimSpace(options.SupervisorName) != "" {
		filtered := make([]DebateRecord, 0, len(records))
		for _, record := range records {
			for _, name := range record.Metadata.ParticipatingSupervisors {
				if strings.EqualFold(name, options.SupervisorName) {
					filtered = append(filtered, record)
					break
				}
			}
		}
		records = filtered
	}
	count, err := s.GetRecordCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	return records, count, nil
}

func (s *DebateHistoryStore) GetStats(ctx context.Context) (*DebateStats, error) {
	records, _, err := s.QueryDebates(ctx, DebateQueryOptions{Limit: 10000})
	if err != nil {
		return nil, err
	}
	stats := &DebateStats{
		DebatesByTaskType:      map[string]int{},
		DebatesBySupervisor:    map[string]int{},
		DebatesByConsensusMode: map[string]int{},
	}
	var totalConsensus float64
	var totalDuration float64
	for _, record := range records {
		stats.TotalDebates++
		if approved, _ := record.Decision["approved"].(bool); approved {
			stats.ApprovedCount++
		} else {
			stats.RejectedCount++
		}
		consensus := floatValueOr(record.Decision["consensus"], 0)
		totalConsensus += consensus
		totalDuration += float64(record.Metadata.DurationMs)
		taskType := stringValue(record.Metadata.DynamicSelection["taskType"])
		if taskType == "" {
			taskType = "general"
		}
		stats.DebatesByTaskType[taskType]++
		stats.DebatesByConsensusMode[record.Metadata.ConsensusMode]++
		for _, supervisor := range record.Metadata.ParticipatingSupervisors {
			stats.DebatesBySupervisor[supervisor]++
		}
		if stats.OldestDebate == nil || record.Timestamp < *stats.OldestDebate {
			copy := record.Timestamp
			stats.OldestDebate = &copy
		}
		if stats.NewestDebate == nil || record.Timestamp > *stats.NewestDebate {
			copy := record.Timestamp
			stats.NewestDebate = &copy
		}
	}
	if stats.TotalDebates > 0 {
		stats.ApprovalRate = float64(stats.ApprovedCount) / float64(stats.TotalDebates)
		stats.AverageConsensus = totalConsensus / float64(stats.TotalDebates)
		stats.AverageDurationMs = totalDuration / float64(stats.TotalDebates)
	}
	return stats, nil
}

func (s *DebateHistoryStore) DeleteRecord(ctx context.Context, id string) (bool, error) {
	db, err := s.open(ctx)
	if err != nil {
		return false, err
	}
	defer db.Close()
	result, err := db.ExecContext(ctx, `DELETE FROM council_debates WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	affected, _ := result.RowsAffected()
	return affected > 0, nil
}

func (s *DebateHistoryStore) GetSupervisorVoteHistory(ctx context.Context, name string) ([]map[string]any, error) {
	records, _, err := s.QueryDebates(ctx, DebateQueryOptions{SupervisorName: name, Limit: 10000})
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0)
	for _, record := range records {
		votes, _ := record.Decision["votes"].([]any)
		for _, vote := range votes {
			voteMap, _ := vote.(map[string]any)
			if !strings.EqualFold(stringValue(voteMap["supervisor"]), name) {
				continue
			}
			result = append(result, map[string]any{
				"debateId":   record.ID,
				"supervisor": stringValue(voteMap["supervisor"]),
				"decision":   stringValue(voteMap["message"]),
				"timestamp":  record.Timestamp,
				"task":       record.Task,
			})
		}
	}
	return result, nil
}

func (s *DebateHistoryStore) ClearAll(ctx context.Context) (int64, error) {
	db, err := s.open(ctx)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	count, err := s.GetRecordCount(ctx)
	if err != nil {
		return 0, err
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM council_debates`); err != nil {
		return 0, err
	}
	return int64(count), nil
}

func (s *DebateHistoryStore) open(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", s.dbPath)
	if err != nil {
		return nil, err
	}
	if err := ensureDebateHistorySchema(ctx, db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func ensureDebateHistorySchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS council_debates (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			session_id TEXT,
			workspace_id TEXT,
			task_type TEXT NOT NULL DEFAULT 'general',
			status TEXT NOT NULL DEFAULT 'completed',
			consensus REAL NOT NULL DEFAULT 0,
			weighted_consensus REAL NOT NULL DEFAULT 0,
			outcome TEXT NOT NULL,
			rounds INTEGER NOT NULL DEFAULT 1,
			timestamp INTEGER NOT NULL,
			data TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_council_debates_session ON council_debates(session_id);
		CREATE INDEX IF NOT EXISTS idx_council_debates_timestamp ON council_debates(timestamp);
	`)
	return err
}

func (s *DebateHistoryStore) prune(ctx context.Context, db *sql.DB) error {
	config := s.GetConfig()
	cutoff := time.Now().UTC().Add(-time.Duration(config.RetentionDays) * 24 * time.Hour).UnixMilli()
	if _, err := db.ExecContext(ctx, `DELETE FROM council_debates WHERE timestamp < ?`, cutoff); err != nil {
		return err
	}
	if config.MaxRecords <= 0 {
		return nil
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM council_debates`).Scan(&count); err != nil {
		return err
	}
	if count <= config.MaxRecords {
		return nil
	}
	excess := count - config.MaxRecords
	_, err := db.ExecContext(ctx, `DELETE FROM council_debates WHERE id IN (SELECT id FROM council_debates ORDER BY timestamp ASC LIMIT ?)`, excess)
	return err
}

func querySingleDebate(ctx context.Context, db *sql.DB, query string, args ...any) (*DebateRecord, error) {
	var raw string
	err := db.QueryRowContext(ctx, query, args...).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var record DebateRecord
	if err := json.Unmarshal([]byte(raw), &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func intValue(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int32:
		return int(typed), true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	default:
		return 0, false
	}
}

func floatValueOr(value any, fallback float64) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	default:
		return fallback
	}
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func truncateString(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) <= max {
		return trimmed
	}
	return trimmed[:max]
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
