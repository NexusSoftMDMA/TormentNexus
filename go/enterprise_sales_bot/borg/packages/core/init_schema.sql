CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT NOT NULL UNIQUE,
            email_verified INTEGER NOT NULL DEFAULT 0,
            image TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
        );
CREATE TABLE IF NOT EXISTS sessions (
            id TEXT PRIMARY KEY,
            expires_at INTEGER NOT NULL,
            token TEXT NOT NULL UNIQUE,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            ip_address TEXT,
            user_agent TEXT,
            user_id TEXT NOT NULL,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
CREATE TABLE IF NOT EXISTS accounts (
            id TEXT PRIMARY KEY,
            account_id TEXT NOT NULL,
            provider_id TEXT NOT NULL,
            user_id TEXT NOT NULL,
            access_token TEXT,
            refresh_token TEXT,
            id_token TEXT,
            access_token_expires_at INTEGER,
            refresh_token_expires_at INTEGER,
            scope TEXT,
            password TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
CREATE TABLE IF NOT EXISTS verifications (
            id TEXT PRIMARY KEY,
            identifier TEXT NOT NULL,
            value TEXT NOT NULL,
            expires_at INTEGER NOT NULL,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
        );
CREATE TABLE IF NOT EXISTS mcp_servers (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            type TEXT NOT NULL DEFAULT 'STDIO',
            command TEXT,
            args TEXT NOT NULL DEFAULT '[]',
            env TEXT NOT NULL DEFAULT '{}',
            url TEXT,
            error_status TEXT NOT NULL DEFAULT 'NONE',
            always_on INTEGER NOT NULL DEFAULT 0,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            bearer_token TEXT,
            headers TEXT NOT NULL DEFAULT '{}',
            user_id TEXT NOT NULL
        );
CREATE TABLE IF NOT EXISTS oauth_sessions (
            uuid TEXT PRIMARY KEY,
            mcp_server_uuid TEXT NOT NULL,
            client_information TEXT NOT NULL DEFAULT '{}',
            tokens TEXT,
            code_verifier TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (mcp_server_uuid) REFERENCES mcp_servers(uuid) ON DELETE CASCADE
        );
CREATE TABLE IF NOT EXISTS tools (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            tool_schema TEXT NOT NULL,
            is_deferred INTEGER NOT NULL DEFAULT 0,
            always_on INTEGER NOT NULL DEFAULT 0,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            mcp_server_uuid TEXT NOT NULL,
            FOREIGN KEY (mcp_server_uuid) REFERENCES mcp_servers(uuid) ON DELETE CASCADE
        );
CREATE TABLE IF NOT EXISTS namespaces (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            user_id TEXT,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
CREATE TABLE IF NOT EXISTS endpoints (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            namespace_uuid TEXT NOT NULL,
            enable_api_key_auth INTEGER NOT NULL DEFAULT 1,
            enable_oauth INTEGER NOT NULL DEFAULT 0,
            enable_max_rate INTEGER NOT NULL DEFAULT 0,
            enable_client_max_rate INTEGER NOT NULL DEFAULT 0,
            max_rate INTEGER,
            max_rate_seconds INTEGER,
            client_max_rate INTEGER,
            client_max_rate_seconds INTEGER,
            client_max_rate_strategy TEXT,
            client_max_rate_strategy_key TEXT,
            use_query_param_auth INTEGER NOT NULL DEFAULT 0,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            user_id TEXT,
            FOREIGN KEY (namespace_uuid) REFERENCES namespaces(uuid) ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
CREATE TABLE IF NOT EXISTS namespace_server_mappings (
            uuid TEXT PRIMARY KEY,
            namespace_uuid TEXT NOT NULL,
            mcp_server_uuid TEXT NOT NULL,
            status TEXT NOT NULL DEFAULT 'ACTIVE',
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (namespace_uuid) REFERENCES namespaces(uuid) ON DELETE CASCADE,
            FOREIGN KEY (mcp_server_uuid) REFERENCES mcp_servers(uuid) ON DELETE CASCADE
        );
CREATE TABLE IF NOT EXISTS namespace_tool_mappings (
            uuid TEXT PRIMARY KEY,
            namespace_uuid TEXT NOT NULL,
            tool_uuid TEXT NOT NULL,
            mcp_server_uuid TEXT NOT NULL,
            status TEXT NOT NULL DEFAULT 'ACTIVE',
            override_name TEXT,
            override_title TEXT,
            override_description TEXT,
            override_annotations TEXT DEFAULT NULL,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (namespace_uuid) REFERENCES namespaces(uuid) ON DELETE CASCADE,
            FOREIGN KEY (tool_uuid) REFERENCES tools(uuid) ON DELETE CASCADE,
            FOREIGN KEY (mcp_server_uuid) REFERENCES mcp_servers(uuid) ON DELETE CASCADE
        );
CREATE TABLE IF NOT EXISTS api_keys (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            key TEXT NOT NULL UNIQUE,
            user_id TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            is_active INTEGER NOT NULL DEFAULT 1,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
CREATE TABLE IF NOT EXISTS config (
            id TEXT PRIMARY KEY,
            value TEXT NOT NULL,
            description TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
        );
CREATE TABLE IF NOT EXISTS oauth_clients (
            client_id TEXT PRIMARY KEY,
            client_secret TEXT,
            client_name TEXT NOT NULL,
            redirect_uris TEXT NOT NULL DEFAULT '[]',
            grant_types TEXT NOT NULL DEFAULT '["authorization_code","refresh_token"]',
            response_types TEXT NOT NULL DEFAULT '["code"]',
            token_endpoint_auth_method TEXT NOT NULL DEFAULT 'none',
            scope TEXT DEFAULT 'admin',
            client_uri TEXT,
            logo_uri TEXT,
            contacts TEXT,
            tos_uri TEXT,
            policy_uri TEXT,
            software_id TEXT,
            software_version TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
        );
CREATE TABLE IF NOT EXISTS oauth_authorization_codes (
            code TEXT PRIMARY KEY,
            client_id TEXT NOT NULL,
            redirect_uri TEXT NOT NULL,
            scope TEXT NOT NULL DEFAULT 'admin',
            user_id TEXT NOT NULL,
            code_challenge TEXT,
            code_challenge_method TEXT,
            expires_at INTEGER NOT NULL,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (client_id) REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
CREATE TABLE IF NOT EXISTS oauth_access_tokens (
            access_token TEXT PRIMARY KEY,
            client_id TEXT NOT NULL,
            user_id TEXT NOT NULL,
            scope TEXT NOT NULL DEFAULT 'admin',
            expires_at INTEGER NOT NULL,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (client_id) REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
CREATE TABLE IF NOT EXISTS docker_sessions (
            uuid TEXT PRIMARY KEY,
            mcp_server_uuid TEXT NOT NULL,
            container_id TEXT NOT NULL,
            container_name TEXT,
            url TEXT,
            status TEXT NOT NULL DEFAULT 'PENDING',
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            started_at INTEGER,
            stopped_at INTEGER,
            error_message TEXT,
            retry_count INTEGER NOT NULL DEFAULT 0,
            last_retry_at INTEGER,
            max_retries INTEGER NOT NULL DEFAULT 3,
            FOREIGN KEY (mcp_server_uuid) REFERENCES mcp_servers(uuid) ON DELETE CASCADE
        );
CREATE TABLE IF NOT EXISTS policies (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            rules TEXT NOT NULL DEFAULT '{}',
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
        );
CREATE TABLE IF NOT EXISTS tool_call_logs (
            uuid TEXT PRIMARY KEY,
            tool_name TEXT NOT NULL,
            mcp_server_uuid TEXT,
            namespace_uuid TEXT,
            endpoint_uuid TEXT,
            args TEXT,
            result TEXT,
            error TEXT,
            duration_ms INTEGER,
            session_id TEXT,
            parent_call_uuid TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (mcp_server_uuid) REFERENCES mcp_servers(uuid) ON DELETE SET NULL,
            FOREIGN KEY (namespace_uuid) REFERENCES namespaces(uuid) ON DELETE SET NULL,
            FOREIGN KEY (endpoint_uuid) REFERENCES endpoints(uuid) ON DELETE SET NULL
        );
CREATE TABLE IF NOT EXISTS tool_sets (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            user_id TEXT,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
CREATE TABLE IF NOT EXISTS tool_set_items (
            uuid TEXT PRIMARY KEY,
            tool_set_uuid TEXT NOT NULL,
            tool_uuid TEXT NOT NULL,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (tool_set_uuid) REFERENCES tool_sets(uuid) ON DELETE CASCADE,
            FOREIGN KEY (tool_uuid) REFERENCES tools(uuid) ON DELETE CASCADE
        );
CREATE TABLE IF NOT EXISTS saved_scripts (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            code TEXT NOT NULL,
            language TEXT NOT NULL DEFAULT 'javascript',
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            user_id TEXT,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );