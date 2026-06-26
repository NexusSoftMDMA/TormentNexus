-- MCP-level health check tracking
ALTER TABLE servers ADD COLUMN health TEXT NOT NULL DEFAULT 'unknown'
    CHECK(health IN ('healthy', 'unhealthy', 'unknown'));
ALTER TABLE servers ADD COLUMN health_check_at DATETIME;
ALTER TABLE servers ADD COLUMN health_error TEXT;
