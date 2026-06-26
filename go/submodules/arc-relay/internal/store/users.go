package store

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                 string    `json:"id"`
	Username           string    `json:"username"`
	PasswordHash       string    `json:"-"`
	Role               string    `json:"role"`
	AccessLevel        string    `json:"access_level"`
	DefaultProfileID   *string   `json:"default_profile_id,omitempty"` // user's default profile for RBAC
	MustChangePassword bool      `json:"must_change_password,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	ProfileID          *string   `json:"profile_id,omitempty"`   // effective profile (resolved at auth time)
	ProfileName        string    `json:"profile_name,omitempty"` // populated on read, not stored
}

type APIKey struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	KeyHash     string     `json:"-"`
	Name        string     `json:"name"`
	ProfileID   *string    `json:"profile_id,omitempty"`
	ProfileName string     `json:"profile_name,omitempty"` // populated on read, not stored
	CreatedAt   time.Time  `json:"created_at"`
	LastUsed    *time.Time `json:"last_used,omitempty"`
	Revoked     bool       `json:"revoked"`
}

type UserStore struct {
	db *DB
}

func NewUserStore(db *DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(username, password, role string) (*User, error) {
	return s.CreateWithAccessLevel(username, password, role, "", nil)
}

func (s *UserStore) CreateWithAccessLevel(username, password, role, accessLevel string, defaultProfileID *string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	// Force admin access level for admin role
	if role == "admin" {
		accessLevel = "admin"
	}
	if accessLevel == "" {
		accessLevel = "write"
	}

	user := &User{
		ID:               uuid.New().String(),
		Username:         username,
		PasswordHash:     string(hash),
		Role:             role,
		AccessLevel:      accessLevel,
		DefaultProfileID: defaultProfileID,
		CreatedAt:        time.Now(),
	}

	_, err = s.db.Exec(`
		INSERT INTO users (id, username, password_hash, role, access_level, default_profile_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Username, user.PasswordHash, user.Role, user.AccessLevel, user.DefaultProfileID, user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return user, nil
}

func (s *UserStore) Authenticate(username, password string) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(`
		SELECT id, username, password_hash, role, access_level, default_profile_id, must_change_password, created_at
		FROM users WHERE username = ?`, username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.AccessLevel, &user.DefaultProfileID, &user.MustChangePassword, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil
	}

	return user, nil
}

func (s *UserStore) Get(id string) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(`
		SELECT id, username, password_hash, role, access_level, default_profile_id, must_change_password, created_at
		FROM users WHERE id = ?`, id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.AccessLevel, &user.DefaultProfileID, &user.MustChangePassword, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}
	return user, nil
}

func (s *UserStore) GetByUsername(username string) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(`
		SELECT id, username, password_hash, role, access_level, default_profile_id, must_change_password, created_at
		FROM users WHERE username = ?`, username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.AccessLevel, &user.DefaultProfileID, &user.MustChangePassword, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by username: %w", err)
	}
	return user, nil
}

func (s *UserStore) List() ([]*User, error) {
	rows, err := s.db.Query(`
		SELECT u.id, u.username, u.password_hash, u.role, u.access_level,
		       u.default_profile_id, u.must_change_password, COALESCE(ap.name, ''), u.created_at
		FROM users u
		LEFT JOIN agent_profiles ap ON u.default_profile_id = ap.id
		ORDER BY u.created_at`)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.AccessLevel,
			&u.DefaultProfileID, &u.MustChangePassword, &u.ProfileName, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *UserStore) UpdateProfile(id string, defaultProfileID *string) error {
	_, err := s.db.Exec(`UPDATE users SET default_profile_id = ? WHERE id = ?`, defaultProfileID, id)
	return err
}

func (s *UserStore) UpdateRole(id, role string) error {
	_, err := s.db.Exec(`UPDATE users SET role = ? WHERE id = ?`, role, id)
	if role == "admin" {
		_, _ = s.db.Exec(`UPDATE users SET access_level = 'admin' WHERE id = ?`, id)
	}
	return err
}

// SetPassword updates a user's password hash and clears the must_change_password flag.
func (s *UserStore) SetPassword(id, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}
	_, err = s.db.Exec(`UPDATE users SET password_hash = ?, must_change_password = FALSE WHERE id = ?`, string(hash), id)
	return err
}

// SetMustChangePassword sets or clears the forced password rotation flag.
func (s *UserStore) SetMustChangePassword(id string, must bool) error {
	_, err := s.db.Exec(`UPDATE users SET must_change_password = ? WHERE id = ?`, must, id)
	return err
}

// CreateWithAccessLevelTx creates a user within an existing transaction.
func (s *UserStore) CreateWithAccessLevelTx(tx *sql.Tx, username, password, role, accessLevel string, defaultProfileID *string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}
	if role == "admin" {
		accessLevel = "admin"
	}
	if accessLevel == "" {
		accessLevel = "write"
	}
	user := &User{
		ID:               uuid.New().String(),
		Username:         username,
		PasswordHash:     string(hash),
		Role:             role,
		AccessLevel:      accessLevel,
		DefaultProfileID: defaultProfileID,
		CreatedAt:        time.Now(),
	}
	_, err = tx.Exec(`
		INSERT INTO users (id, username, password_hash, role, access_level, default_profile_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Username, user.PasswordHash, user.Role, user.AccessLevel, user.DefaultProfileID, user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return user, nil
}

// CreateAPIKeyTx generates a new API key within an existing transaction.
func (s *UserStore) CreateAPIKeyTx(tx *sql.Tx, userID, name string, profileID *string) (string, *APIKey, error) {
	rawKey := uuid.New().String()
	keyHash := hashAPIKey(rawKey)
	ak := &APIKey{
		ID:        uuid.New().String(),
		UserID:    userID,
		KeyHash:   keyHash,
		Name:      name,
		ProfileID: profileID,
		CreatedAt: time.Now(),
	}
	_, err := tx.Exec(`
		INSERT INTO api_keys (id, user_id, key_hash, name, profile_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		ak.ID, ak.UserID, ak.KeyHash, ak.Name, ak.ProfileID, ak.CreatedAt,
	)
	if err != nil {
		return "", nil, fmt.Errorf("creating api key: %w", err)
	}
	return rawKey, ak, nil
}

func (s *UserStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

// EnsureAdmin creates the default admin user if no users exist.
// Also ensures existing admin users have access_level = 'admin'.
func (s *UserStore) EnsureAdmin(password string) error {
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return fmt.Errorf("counting users: %w", err)
	}
	if count > 0 {
		// Ensure all admin-role users have admin access level
		_, _ = s.db.Exec(`UPDATE users SET access_level = 'admin' WHERE role = 'admin' AND access_level != 'admin'`)
		return nil
	}
	_, err := s.Create("admin", password, "admin")
	return err
}

// API Key operations

func hashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// CreateAPIKey generates a new API key and returns it (plaintext shown once).
func (s *UserStore) CreateAPIKey(userID, name string, profileID *string) (string, *APIKey, error) {
	rawKey := uuid.New().String() // the plaintext key
	keyHash := hashAPIKey(rawKey)

	ak := &APIKey{
		ID:        uuid.New().String(),
		UserID:    userID,
		KeyHash:   keyHash,
		Name:      name,
		ProfileID: profileID,
		CreatedAt: time.Now(),
	}

	_, err := s.db.Exec(`
		INSERT INTO api_keys (id, user_id, key_hash, name, profile_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		ak.ID, ak.UserID, ak.KeyHash, ak.Name, ak.ProfileID, ak.CreatedAt,
	)
	if err != nil {
		return "", nil, fmt.Errorf("creating api key: %w", err)
	}
	return rawKey, ak, nil
}

// ValidateAPIKey checks a raw API key and returns the associated user.
// Resolution order for effective profile:
//  1. Key has explicit profile_id → use that
//  2. Owning user has default_profile_id → use that
//  3. No profile → legacy tier-based access via access_level
func (s *UserStore) ValidateAPIKey(rawKey string) (*User, error) {
	keyHash := hashAPIKey(rawKey)

	var userID string
	var storedHash string
	var revoked bool
	var keyProfileID sql.NullString
	err := s.db.QueryRow(`
		SELECT user_id, key_hash, revoked, profile_id FROM api_keys WHERE key_hash = ?`, keyHash,
	).Scan(&userID, &storedHash, &revoked, &keyProfileID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up api key: %w", err)
	}

	// Constant-time comparison
	if subtle.ConstantTimeCompare([]byte(keyHash), []byte(storedHash)) != 1 {
		return nil, nil
	}
	if revoked {
		return nil, nil
	}

	// Update last_used
	_, _ = s.db.Exec("UPDATE api_keys SET last_used = ? WHERE key_hash = ?", time.Now(), keyHash)

	user, err := s.Get(userID)
	if err != nil || user == nil {
		return nil, err
	}

	// Resolve effective profile: key-level override > user default > none
	if keyProfileID.Valid {
		user.ProfileID = &keyProfileID.String
	} else if user.DefaultProfileID != nil {
		user.ProfileID = user.DefaultProfileID
	}

	return user, nil
}

func (s *UserStore) ListAPIKeys(userID string) ([]*APIKey, error) {
	rows, err := s.db.Query(`
		SELECT ak.id, ak.user_id, ak.name, ak.profile_id, COALESCE(ap.name, ''), ak.created_at, ak.last_used, ak.revoked
		FROM api_keys ak
		LEFT JOIN agent_profiles ap ON ak.profile_id = ap.id
		WHERE ak.user_id = ? ORDER BY ak.created_at`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing api keys: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var keys []*APIKey
	for rows.Next() {
		k := &APIKey{}
		var profileID sql.NullString
		if err := rows.Scan(&k.ID, &k.UserID, &k.Name, &profileID, &k.ProfileName, &k.CreatedAt, &k.LastUsed, &k.Revoked); err != nil {
			return nil, fmt.Errorf("scanning api key: %w", err)
		}
		if profileID.Valid {
			k.ProfileID = &profileID.String
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *UserStore) RevokeAPIKey(id string) error {
	_, err := s.db.Exec("UPDATE api_keys SET revoked = TRUE WHERE id = ?", id)
	return err
}
