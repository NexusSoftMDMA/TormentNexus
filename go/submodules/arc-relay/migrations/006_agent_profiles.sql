-- Agent profiles for granular per-operation RBAC
CREATE TABLE IF NOT EXISTS agent_profiles (
    id          TEXT PRIMARY KEY,
    name        TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS profile_permissions (
    profile_id    TEXT NOT NULL REFERENCES agent_profiles(id) ON DELETE CASCADE,
    server_id     TEXT NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    endpoint_type TEXT NOT NULL CHECK(endpoint_type IN ('tool', 'resource', 'prompt')),
    endpoint_name TEXT NOT NULL,
    PRIMARY KEY (profile_id, server_id, endpoint_type, endpoint_name)
);

CREATE INDEX IF NOT EXISTS idx_profile_perms_profile ON profile_permissions(profile_id);
CREATE INDEX IF NOT EXISTS idx_profile_perms_server ON profile_permissions(server_id);

ALTER TABLE api_keys ADD COLUMN profile_id TEXT REFERENCES agent_profiles(id) ON DELETE SET NULL;
