-- Invite tokens allow admins to onboard users without requiring web UI login.
-- A token is created for a specific user, and when exchanged via the CLI,
-- creates an API key for that user with the specified profile.
CREATE TABLE IF NOT EXISTS invite_tokens (
    id          TEXT PRIMARY KEY,
    token_hash  TEXT UNIQUE NOT NULL,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_id  TEXT REFERENCES agent_profiles(id) ON DELETE CASCADE,
    created_by  TEXT NOT NULL REFERENCES users(id),
    expires_at  DATETIME NOT NULL,
    used_at     DATETIME,
    status      TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'used', 'expired'))
);
