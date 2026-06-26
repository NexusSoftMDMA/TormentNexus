package store

import (
	"time"
)

// SessionStore persists web UI sessions to SQLite so they survive restarts.
type SessionStore struct {
	db *DB
}

func NewSessionStore(db *DB) *SessionStore {
	return &SessionStore{db: db}
}

// Create stores a new session.
func (s *SessionStore) Create(id, userID string, expiresAt time.Time) error {
	_, err := s.db.Exec(
		`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`,
		id, userID, expiresAt,
	)
	return err
}

// Get returns the user for a valid (non-expired) session, or nil.
// Loads full user data including access_level and default_profile_id for authorization.
func (s *SessionStore) Get(id string) (*User, time.Time, bool) {
	var user User
	var expiresAt time.Time
	err := s.db.QueryRow(`
		SELECT u.id, u.username, u.role, u.access_level, u.default_profile_id, u.must_change_password, s.expires_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.id = ? AND s.expires_at > ?`,
		id, time.Now(),
	).Scan(&user.ID, &user.Username, &user.Role, &user.AccessLevel, &user.DefaultProfileID, &user.MustChangePassword, &expiresAt)
	if err != nil {
		return nil, time.Time{}, false
	}
	// For web sessions, the effective profile is the user's default profile
	user.ProfileID = user.DefaultProfileID
	return &user, expiresAt, true
}

// Delete removes a session (logout).
func (s *SessionStore) Delete(id string) {
	_, _ = s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
}

// DeleteByUser removes all sessions for a specific user (e.g., after password reset).
func (s *SessionStore) DeleteByUser(userID string) {
	_, _ = s.db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
}

// Cleanup removes all expired sessions.
func (s *SessionStore) Cleanup() {
	_, _ = s.db.Exec(`DELETE FROM sessions WHERE expires_at < ?`, time.Now())
}
