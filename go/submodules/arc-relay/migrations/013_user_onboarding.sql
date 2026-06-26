-- Unified onboarding: invites become account templates, users set their own credentials.
-- Also adds must_change_password for admin-forced password rotation.

-- 1. Add forced password change flag to users.
ALTER TABLE users ADD COLUMN must_change_password BOOLEAN NOT NULL DEFAULT FALSE;

-- 2. Recreate invite_tokens without user_id, with account-template columns.
CREATE TABLE invite_tokens_new (
    id               TEXT PRIMARY KEY,
    token_hash       TEXT UNIQUE NOT NULL,
    role             TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('admin', 'user')),
    access_level     TEXT NOT NULL DEFAULT 'write' CHECK(access_level IN ('read', 'write', 'admin')),
    profile_id       TEXT REFERENCES agent_profiles(id) ON DELETE SET NULL,
    created_by       TEXT NOT NULL REFERENCES users(id),
    expires_at       DATETIME NOT NULL,
    used_at          DATETIME,
    redeemed_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    status           TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'used', 'expired'))
);

-- Preserve history from old table; expire any pending invites (breaking change).
INSERT INTO invite_tokens_new (id, token_hash, role, access_level, profile_id, created_by, expires_at, used_at, redeemed_user_id, status)
SELECT it.id, it.token_hash,
       COALESCE((SELECT role FROM users WHERE id = it.user_id), 'user'),
       COALESCE((SELECT access_level FROM users WHERE id = it.user_id), 'write'),
       it.profile_id, it.created_by, it.expires_at, it.used_at,
       CASE WHEN it.status = 'used' THEN it.user_id ELSE NULL END,
       CASE WHEN it.status = 'pending' THEN 'expired' ELSE it.status END
FROM invite_tokens it;

DROP TABLE invite_tokens;
ALTER TABLE invite_tokens_new RENAME TO invite_tokens;
