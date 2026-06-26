package store

import (
	"fmt"
	"log/slog"
)

// EndpointTier represents the access tier for a single endpoint.
type EndpointTier struct {
	ServerID       string `json:"server_id"`
	EndpointType   string `json:"endpoint_type"`
	EndpointName   string `json:"endpoint_name"`
	AccessTier     string `json:"access_tier"`
	AutoClassified bool   `json:"auto_classified"`
}

// AccessStore manages endpoint access tiers and user access level checks.
type AccessStore struct {
	db *DB
}

func NewAccessStore(db *DB) *AccessStore {
	return &AccessStore{db: db}
}

// GetTier returns the access tier for an endpoint, or "write" as default.
func (s *AccessStore) GetTier(serverID, endpointType, endpointName string) string {
	var tier string
	err := s.db.QueryRow(`
		SELECT access_tier FROM endpoint_access_tiers
		WHERE server_id = ? AND endpoint_type = ? AND endpoint_name = ?`,
		serverID, endpointType, endpointName,
	).Scan(&tier)
	if err != nil {
		return "write"
	}
	return tier
}

// GetAllTiers returns all access tiers for a server.
func (s *AccessStore) GetAllTiers(serverID string) ([]EndpointTier, error) {
	rows, err := s.db.Query(`
		SELECT server_id, endpoint_type, endpoint_name, access_tier, auto_classified
		FROM endpoint_access_tiers WHERE server_id = ?
		ORDER BY endpoint_type, endpoint_name`, serverID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing access tiers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tiers []EndpointTier
	for rows.Next() {
		var t EndpointTier
		if err := rows.Scan(&t.ServerID, &t.EndpointType, &t.EndpointName, &t.AccessTier, &t.AutoClassified); err != nil {
			return nil, fmt.Errorf("scanning access tier: %w", err)
		}
		tiers = append(tiers, t)
	}
	return tiers, nil
}

// SetTier sets the access tier for an endpoint, marking it as manually overridden.
func (s *AccessStore) SetTier(serverID, endpointType, endpointName, tier string) error {
	_, err := s.db.Exec(`
		INSERT INTO endpoint_access_tiers (server_id, endpoint_type, endpoint_name, access_tier, auto_classified)
		VALUES (?, ?, ?, ?, FALSE)
		ON CONFLICT(server_id, endpoint_type, endpoint_name)
		DO UPDATE SET access_tier = ?, auto_classified = FALSE`,
		serverID, endpointType, endpointName, tier, tier,
	)
	return err
}

// EndpointInfo is a minimal struct for SyncAfterEnumerate input.
type EndpointInfo struct {
	Type        string
	Name        string
	Description string
}

// SyncAfterEnumerate upserts auto-classified tiers for newly discovered endpoints,
// preserves manual overrides (auto_classified = FALSE), and deletes stale entries.
func (s *AccessStore) SyncAfterEnumerate(serverID string, endpoints []EndpointInfo, classifyFunc func(endpointType, name, description string) string) {
	tx, err := s.db.Begin()
	if err != nil {
		slog.Warn("SyncAfterEnumerate: begin tx failed", "err", err)
		return
	}

	// Build set of current endpoints
	current := make(map[string]bool) // "type:name" -> true
	for _, ep := range endpoints {
		key := ep.Type + ":" + ep.Name
		current[key] = true

		tier := classifyFunc(ep.Type, ep.Name, ep.Description)

		// Upsert: insert with auto-classified tier, but only update if auto_classified = TRUE
		_, err := tx.Exec(`
			INSERT INTO endpoint_access_tiers (server_id, endpoint_type, endpoint_name, access_tier, auto_classified)
			VALUES (?, ?, ?, ?, TRUE)
			ON CONFLICT(server_id, endpoint_type, endpoint_name)
			DO UPDATE SET access_tier = ?, auto_classified = TRUE
			WHERE auto_classified = TRUE`,
			serverID, ep.Type, ep.Name, tier, tier,
		)
		if err != nil {
			slog.Warn("SyncAfterEnumerate: upsert failed", "type", ep.Type, "name", ep.Name, "err", err)
		}
	}

	// Delete stale entries (endpoints no longer present)
	existing, err := s.GetAllTiers(serverID)
	if err == nil {
		for _, t := range existing {
			key := t.EndpointType + ":" + t.EndpointName
			if !current[key] {
				_, _ = tx.Exec(`DELETE FROM endpoint_access_tiers
					WHERE server_id = ? AND endpoint_type = ? AND endpoint_name = ?`,
					serverID, t.EndpointType, t.EndpointName)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Warn("SyncAfterEnumerate: commit failed", "err", err)
	}
}

// tierLevel converts a tier string to a numeric level for comparison.
func tierLevel(tier string) int {
	switch tier {
	case "read":
		return 1
	case "write":
		return 2
	case "admin":
		return 3
	default:
		return 2 // default to write
	}
}

// CheckAccess returns true if the user's access level is sufficient for the endpoint tier.
func (s *AccessStore) CheckAccess(userLevel, endpointTier string) bool {
	return tierLevel(userLevel) >= tierLevel(endpointTier)
}
