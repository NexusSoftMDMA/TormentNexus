package store

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AgentProfile defines a named set of endpoint permissions across servers.
type AgentProfile struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProfilePermission grants access to a single endpoint on a server.
type ProfilePermission struct {
	ProfileID    string `json:"profile_id"`
	ServerID     string `json:"server_id"`
	EndpointType string `json:"endpoint_type"`
	EndpointName string `json:"endpoint_name"`
}

// ProfileStore manages agent profiles and their permissions.
type ProfileStore struct {
	db *DB
}

func NewProfileStore(db *DB) *ProfileStore {
	return &ProfileStore{db: db}
}

func (s *ProfileStore) Create(name, description string) (*AgentProfile, error) {
	p := &AgentProfile{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_, err := s.db.Exec(`
		INSERT INTO agent_profiles (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Description, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating profile: %w", err)
	}
	return p, nil
}

func (s *ProfileStore) Get(id string) (*AgentProfile, error) {
	p := &AgentProfile{}
	err := s.db.QueryRow(`
		SELECT id, name, description, created_at, updated_at
		FROM agent_profiles WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting profile: %w", err)
	}
	return p, nil
}

func (s *ProfileStore) GetByName(name string) (*AgentProfile, error) {
	p := &AgentProfile{}
	err := s.db.QueryRow(`
		SELECT id, name, description, created_at, updated_at
		FROM agent_profiles WHERE name = ?`, name,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting profile by name: %w", err)
	}
	return p, nil
}

func (s *ProfileStore) List() ([]*AgentProfile, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, created_at, updated_at
		FROM agent_profiles ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("listing profiles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var profiles []*AgentProfile
	for rows.Next() {
		p := &AgentProfile{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning profile: %w", err)
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

func (s *ProfileStore) Update(id, name, description string) error {
	_, err := s.db.Exec(`
		UPDATE agent_profiles SET name = ?, description = ?, updated_at = ?
		WHERE id = ?`,
		name, description, time.Now(), id,
	)
	return err
}

func (s *ProfileStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM agent_profiles WHERE id = ?", id)
	return err
}

// GetPermissions returns all permissions for a profile.
func (s *ProfileStore) GetPermissions(profileID string) ([]ProfilePermission, error) {
	rows, err := s.db.Query(`
		SELECT profile_id, server_id, endpoint_type, endpoint_name
		FROM profile_permissions WHERE profile_id = ?
		ORDER BY server_id, endpoint_type, endpoint_name`, profileID,
	)
	if err != nil {
		return nil, fmt.Errorf("getting permissions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var perms []ProfilePermission
	for rows.Next() {
		var p ProfilePermission
		if err := rows.Scan(&p.ProfileID, &p.ServerID, &p.EndpointType, &p.EndpointName); err != nil {
			return nil, fmt.Errorf("scanning permission: %w", err)
		}
		perms = append(perms, p)
	}
	return perms, nil
}

// GetPermissionsForServer returns permissions for a single server.
func (s *ProfileStore) GetPermissionsForServer(profileID, serverID string) ([]ProfilePermission, error) {
	rows, err := s.db.Query(`
		SELECT profile_id, server_id, endpoint_type, endpoint_name
		FROM profile_permissions WHERE profile_id = ? AND server_id = ?
		ORDER BY endpoint_type, endpoint_name`, profileID, serverID,
	)
	if err != nil {
		return nil, fmt.Errorf("getting server permissions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var perms []ProfilePermission
	for rows.Next() {
		var p ProfilePermission
		if err := rows.Scan(&p.ProfileID, &p.ServerID, &p.EndpointType, &p.EndpointName); err != nil {
			return nil, fmt.Errorf("scanning permission: %w", err)
		}
		perms = append(perms, p)
	}
	return perms, nil
}

// SetPermission grants access to a single endpoint.
func (s *ProfileStore) SetPermission(profileID, serverID, endpointType, endpointName string) error {
	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO profile_permissions (profile_id, server_id, endpoint_type, endpoint_name)
		VALUES (?, ?, ?, ?)`,
		profileID, serverID, endpointType, endpointName,
	)
	return err
}

// RemovePermission revokes access to a single endpoint.
func (s *ProfileStore) RemovePermission(profileID, serverID, endpointType, endpointName string) error {
	_, err := s.db.Exec(`
		DELETE FROM profile_permissions
		WHERE profile_id = ? AND server_id = ? AND endpoint_type = ? AND endpoint_name = ?`,
		profileID, serverID, endpointType, endpointName,
	)
	return err
}

// BulkSetPermissions replaces all permissions for a profile+server combination.
func (s *ProfileStore) BulkSetPermissions(profileID, serverID string, perms []ProfilePermission) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// Clear existing permissions for this profile+server
	if _, err := tx.Exec(`
		DELETE FROM profile_permissions WHERE profile_id = ? AND server_id = ?`,
		profileID, serverID,
	); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("clearing permissions: %w", err)
	}

	// Insert new permissions
	for _, p := range perms {
		if _, err := tx.Exec(`
			INSERT INTO profile_permissions (profile_id, server_id, endpoint_type, endpoint_name)
			VALUES (?, ?, ?, ?)`,
			profileID, serverID, p.EndpointType, p.EndpointName,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("inserting permission: %w", err)
		}
	}

	return tx.Commit()
}

// CheckPermission returns true if the profile grants access to the endpoint.
func (s *ProfileStore) CheckPermission(profileID, serverID, endpointType, endpointName string) (bool, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM profile_permissions
		WHERE profile_id = ? AND server_id = ? AND endpoint_type = ? AND endpoint_name = ?`,
		profileID, serverID, endpointType, endpointName,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking permission: %w", err)
	}
	return count > 0, nil
}

// SeedFromTier bulk-inserts permissions from endpoint_access_tiers for a given max tier level.
// maxTier should be "read", "write", or "admin".
func (s *ProfileStore) SeedFromTier(profileID, serverID, maxTier string) error {
	var condition string
	switch maxTier {
	case "read":
		condition = "access_tier = 'read'"
	case "write":
		condition = "access_tier IN ('read', 'write')"
	case "admin":
		condition = "1=1"
	default:
		return fmt.Errorf("invalid tier: %s", maxTier)
	}

	_, err := s.db.Exec(fmt.Sprintf(`
		INSERT OR IGNORE INTO profile_permissions (profile_id, server_id, endpoint_type, endpoint_name)
		SELECT ?, server_id, endpoint_type, endpoint_name
		FROM endpoint_access_tiers
		WHERE server_id = ? AND %s`, condition),
		profileID, serverID,
	)
	return err
}

// PermissionCount returns the number of permissions for a profile.
func (s *ProfileStore) PermissionCount(profileID string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM profile_permissions WHERE profile_id = ?`, profileID,
	).Scan(&count)
	return count, err
}

// APIKeyCount returns the number of API keys assigned to a profile.
func (s *ProfileStore) APIKeyCount(profileID string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM api_keys WHERE profile_id = ? AND revoked = FALSE`, profileID,
	).Scan(&count)
	return count, err
}

// ServerIDsForProfile returns distinct server IDs that the profile has any permissions for,
// filtered to only servers that still exist.
func (s *ProfileStore) ServerIDsForProfile(profileID string) (map[string]bool, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT pp.server_id
		FROM profile_permissions pp
		JOIN servers s ON s.id = pp.server_id
		WHERE pp.profile_id = ?`, profileID,
	)
	if err != nil {
		return nil, fmt.Errorf("getting server IDs for profile: %w", err)
	}
	defer func() { _ = rows.Close() }()

	ids := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning server ID: %w", err)
		}
		ids[id] = true
	}
	return ids, rows.Err()
}
