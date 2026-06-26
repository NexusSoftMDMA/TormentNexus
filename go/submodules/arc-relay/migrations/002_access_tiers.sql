-- Endpoint access tiers and user access levels

DROP TABLE IF EXISTS endpoint_permissions;  -- unused, replacing with tier model

CREATE TABLE IF NOT EXISTS endpoint_access_tiers (
    server_id       TEXT NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    endpoint_type   TEXT NOT NULL CHECK(endpoint_type IN ('tool', 'resource', 'prompt')),
    endpoint_name   TEXT NOT NULL,
    access_tier     TEXT NOT NULL DEFAULT 'write' CHECK(access_tier IN ('read', 'write', 'admin')),
    auto_classified BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (server_id, endpoint_type, endpoint_name)
);
CREATE INDEX IF NOT EXISTS idx_access_tiers_server ON endpoint_access_tiers(server_id);

ALTER TABLE users ADD COLUMN access_level TEXT NOT NULL DEFAULT 'write'
    CHECK(access_level IN ('read', 'write', 'admin'));
