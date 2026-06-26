package store

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RequestLog struct {
	ID           string
	Timestamp    time.Time
	UserID       string
	Username     string // joined for display
	ServerID     string
	ServerName   string // joined for display
	Method       string
	EndpointName string
	DurationMs   int64
	Status       string // "success", "error", "denied"
	ErrorMsg     string
}

type LogStats struct {
	TotalRequests24h int
	Errors24h        int
	AvgDurationMs    int
	ByServer         []ServerRequestCount
}

type ServerRequestCount struct {
	ServerName string
	Count      int
}

type EndpointCallCount struct {
	EndpointName  string
	CallCount     int
	ErrorCount    int
	AvgDurationMs int
}

type RequestLogStore struct {
	db *DB
}

func NewRequestLogStore(db *DB) *RequestLogStore {
	return &RequestLogStore{db: db}
}

func (s *RequestLogStore) Create(rl *RequestLog) error {
	if rl.ID == "" {
		rl.ID = uuid.New().String()
	}
	if rl.Timestamp.IsZero() {
		rl.Timestamp = time.Now()
	}

	_, err := s.db.Exec(`
		INSERT INTO request_logs (id, timestamp, user_id, server_id, method, endpoint_name, duration_ms, status, error_msg)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rl.ID, rl.Timestamp, rl.UserID, rl.ServerID, rl.Method, rl.EndpointName, rl.DurationMs, rl.Status, rl.ErrorMsg,
	)
	if err != nil {
		return fmt.Errorf("creating request log: %w", err)
	}
	return nil
}

func (s *RequestLogStore) Recent(limit int) ([]*RequestLog, error) {
	rows, err := s.db.Query(`
		SELECT rl.id, rl.timestamp, rl.user_id, COALESCE(u.username, ''), rl.server_id, COALESCE(sv.name, ''),
		       rl.method, COALESCE(rl.endpoint_name, ''), COALESCE(rl.duration_ms, 0), COALESCE(rl.status, ''), COALESCE(rl.error_msg, '')
		FROM request_logs rl
		LEFT JOIN users u ON rl.user_id = u.id
		LEFT JOIN servers sv ON rl.server_id = sv.id
		ORDER BY rl.timestamp DESC
		LIMIT ?`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("querying recent logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanLogs(rows)
}

func (s *RequestLogStore) ByServer(serverID string, limit int) ([]*RequestLog, error) {
	rows, err := s.db.Query(`
		SELECT rl.id, rl.timestamp, rl.user_id, COALESCE(u.username, ''), rl.server_id, COALESCE(sv.name, ''),
		       rl.method, COALESCE(rl.endpoint_name, ''), COALESCE(rl.duration_ms, 0), COALESCE(rl.status, ''), COALESCE(rl.error_msg, '')
		FROM request_logs rl
		LEFT JOIN users u ON rl.user_id = u.id
		LEFT JOIN servers sv ON rl.server_id = sv.id
		WHERE rl.server_id = ?
		ORDER BY rl.timestamp DESC
		LIMIT ?`, serverID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("querying server logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanLogs(rows)
}

func (s *RequestLogStore) Stats() (*LogStats, error) {
	stats := &LogStats{}

	// Total requests and errors in last 24h
	err := s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(CASE WHEN status='error' THEN 1 ELSE 0 END), 0),
		       COALESCE(AVG(duration_ms), 0)
		FROM request_logs
		WHERE timestamp >= datetime('now', '-24 hours')`,
	).Scan(&stats.TotalRequests24h, &stats.Errors24h, &stats.AvgDurationMs)
	if err != nil {
		return stats, fmt.Errorf("querying log stats: %w", err)
	}

	// Requests by server in last 24h
	rows, err := s.db.Query(`
		SELECT COALESCE(sv.name, 'unknown'), COUNT(*)
		FROM request_logs rl
		LEFT JOIN servers sv ON rl.server_id = sv.id
		WHERE rl.timestamp >= datetime('now', '-24 hours')
		GROUP BY rl.server_id
		ORDER BY COUNT(*) DESC`,
	)
	if err != nil {
		return stats, nil
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var sc ServerRequestCount
		if err := rows.Scan(&sc.ServerName, &sc.Count); err != nil {
			continue
		}
		stats.ByServer = append(stats.ByServer, sc)
	}

	return stats, nil
}

func (s *RequestLogStore) EndpointCounts(serverID string) ([]EndpointCallCount, error) {
	rows, err := s.db.Query(`
		SELECT endpoint_name, COUNT(*), COUNT(CASE WHEN status='error' THEN 1 END), COALESCE(AVG(duration_ms), 0)
		FROM request_logs
		WHERE server_id = ? AND endpoint_name IS NOT NULL AND endpoint_name != ''
		GROUP BY endpoint_name`, serverID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying endpoint counts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var counts []EndpointCallCount
	for rows.Next() {
		var ec EndpointCallCount
		if err := rows.Scan(&ec.EndpointName, &ec.CallCount, &ec.ErrorCount, &ec.AvgDurationMs); err != nil {
			continue
		}
		counts = append(counts, ec)
	}
	return counts, nil
}

// LogFilter holds optional filters and pagination for querying logs.
type LogFilter struct {
	ServerID string
	UserID   string
	Status   string // "success", "error", "denied", or "" for all
	Limit    int
	Offset   int
}

// FilteredLogs returns logs matching the given filters with pagination.
func (s *RequestLogStore) FilteredLogs(f LogFilter) ([]*RequestLog, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}

	where := "1=1"
	args := []any{}

	if f.ServerID != "" {
		where += " AND rl.server_id = ?"
		args = append(args, f.ServerID)
	}
	if f.UserID != "" {
		where += " AND rl.user_id = ?"
		args = append(args, f.UserID)
	}
	if f.Status != "" {
		where += " AND rl.status = ?"
		args = append(args, f.Status)
	}

	// Count total matching rows
	var total int
	countQuery := "SELECT COUNT(*) FROM request_logs rl WHERE " + where
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting logs: %w", err)
	}

	// Fetch page
	query := `
		SELECT rl.id, rl.timestamp, rl.user_id, COALESCE(u.username, ''), rl.server_id, COALESCE(sv.name, ''),
		       rl.method, COALESCE(rl.endpoint_name, ''), COALESCE(rl.duration_ms, 0), COALESCE(rl.status, ''), COALESCE(rl.error_msg, '')
		FROM request_logs rl
		LEFT JOIN users u ON rl.user_id = u.id
		LEFT JOIN servers sv ON rl.server_id = sv.id
		WHERE ` + where + `
		ORDER BY rl.timestamp DESC
		LIMIT ? OFFSET ?`
	args = append(args, f.Limit, f.Offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying filtered logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	logs, err := scanLogs(rows)
	return logs, total, err
}

// DistinctUsers returns usernames that appear in the logs (for filter dropdown).
func (s *RequestLogStore) DistinctUsers() ([]struct{ ID, Username string }, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT rl.user_id, COALESCE(u.username, '')
		FROM request_logs rl
		LEFT JOIN users u ON rl.user_id = u.id
		WHERE rl.user_id != ''
		ORDER BY u.username`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var users []struct{ ID, Username string }
	for rows.Next() {
		var u struct{ ID, Username string }
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *RequestLogStore) ServerTotalCounts() (map[string]int, error) {
	rows, err := s.db.Query(`SELECT server_id, COUNT(*) FROM request_logs GROUP BY server_id`)
	if err != nil {
		return nil, fmt.Errorf("querying server total counts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	counts := make(map[string]int)
	for rows.Next() {
		var serverID string
		var count int
		if err := rows.Scan(&serverID, &count); err != nil {
			continue
		}
		counts[serverID] = count
	}
	return counts, nil
}

func scanLogs(rows interface {
	Next() bool
	Scan(dest ...any) error
}) ([]*RequestLog, error) {
	var logs []*RequestLog
	for rows.Next() {
		rl := &RequestLog{}
		if err := rows.Scan(
			&rl.ID, &rl.Timestamp, &rl.UserID, &rl.Username, &rl.ServerID, &rl.ServerName,
			&rl.Method, &rl.EndpointName, &rl.DurationMs, &rl.Status, &rl.ErrorMsg,
		); err != nil {
			return nil, fmt.Errorf("scanning request log: %w", err)
		}
		logs = append(logs, rl)
	}
	return logs, nil
}
