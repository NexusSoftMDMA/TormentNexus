-- Add default_profile_id to users, linking users to profiles for RBAC.
-- API keys inherit the user's default profile when the key has no explicit profile.
ALTER TABLE users ADD COLUMN default_profile_id TEXT REFERENCES agent_profiles(id) ON DELETE SET NULL;
