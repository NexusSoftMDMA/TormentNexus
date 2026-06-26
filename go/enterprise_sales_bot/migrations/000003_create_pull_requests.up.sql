-- 000003_create_pull_requests.up.sql
CREATE TABLE IF NOT EXISTS pull_requests (
    id TEXT PRIMARY KEY,
    branch TEXT NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL,
    task_description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
