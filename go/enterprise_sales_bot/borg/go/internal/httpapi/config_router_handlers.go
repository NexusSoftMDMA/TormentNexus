package httpapi

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

func (s *Server) handleConfigList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "config.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "config.list",
			},
		})
		return
	}

	configs, fallbackErr := s.localConfigList()
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    configs,
		"bridge": map[string]any{
			"fallback":  "go-local-config-db",
			"procedure": "config.list",
			"reason":    "upstream unavailable; using local tormentnexus config table",
		},
	})
}

func (s *Server) handleConfigGet(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.URL.Query().Get("key"))
	if key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing key query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "config.get", map[string]any{"key": key}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "config.get",
			},
		})
		return
	}

	value, fallbackErr := s.localConfigValue(key)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    value,
		"bridge": map[string]any{
			"fallback":  "go-local-config-db",
			"procedure": "config.get",
			"reason":    "upstream unavailable; using local tormentnexus config value",
		},
	})
}

func (s *Server) handleConfigUpsert(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "config.upsert")
}

func (s *Server) handleConfigDelete(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "config.delete")
}

func (s *Server) handleConfigUpdate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "config.update")
}

func (s *Server) handleConfigGetMCPTimeout(w http.ResponseWriter, r *http.Request) {
	s.handleConfigScalarFallback(w, r, "config.getMcpTimeout", "MCP_TIMEOUT", 60000)
}

func (s *Server) handleConfigSetMCPTimeout(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "config.setMcpTimeout")
}

func (s *Server) handleConfigGetMCPMaxAttempts(w http.ResponseWriter, r *http.Request) {
	s.handleConfigScalarFallback(w, r, "config.getMcpMaxAttempts", "MCP_MAX_ATTEMPTS", 1)
}

func (s *Server) handleConfigSetMCPMaxAttempts(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "config.setMcpMaxAttempts")
}

func (s *Server) handleConfigGetMCPMaxTotalTimeout(w http.ResponseWriter, r *http.Request) {
	s.handleConfigScalarFallback(w, r, "config.getMcpMaxTotalTimeout", "MCP_MAX_TOTAL_TIMEOUT", 60000)
}

func (s *Server) handleConfigSetMCPMaxTotalTimeout(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "config.setMcpMaxTotalTimeout")
}

func (s *Server) handleConfigGetMCPResetTimeoutOnProgress(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "config.getMcpResetTimeoutOnProgress", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "config.getMcpResetTimeoutOnProgress",
			},
		})
		return
	}

	// Preserve current TS behavior: config === "true" || true always returns true.
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    true,
		"bridge": map[string]any{
			"fallback":  "go-local-config-db",
			"procedure": "config.getMcpResetTimeoutOnProgress",
			"reason":    "upstream unavailable; preserving current config service default semantics",
		},
	})
}

func (s *Server) handleConfigSetMCPResetTimeoutOnProgress(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "config.setMcpResetTimeoutOnProgress")
}

func (s *Server) handleConfigGetSessionLifetime(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "config.getSessionLifetime", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "config.getSessionLifetime",
			},
		})
		return
	}

	value, fallbackErr := s.localConfigValue("SESSION_LIFETIME")
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}
	if value == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    nil,
			"bridge": map[string]any{
				"fallback":  "go-local-config-db",
				"procedure": "config.getSessionLifetime",
				"reason":    "upstream unavailable; using local tormentnexus session lifetime config",
			},
		})
		return
	}

	lifetime, _ := strconv.Atoi(value.(string))
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    lifetime,
		"bridge": map[string]any{
			"fallback":  "go-local-config-db",
			"procedure": "config.getSessionLifetime",
			"reason":    "upstream unavailable; using local tormentnexus session lifetime config",
		},
	})
}

func (s *Server) handleConfigSetSessionLifetime(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "config.setSessionLifetime")
}

func (s *Server) handleConfigGetSignupDisabled(w http.ResponseWriter, r *http.Request) {
	s.handleConfigBooleanFallback(w, r, "config.getSignupDisabled", "DISABLE_SIGNUP", false)
}

func (s *Server) handleConfigSetSignupDisabled(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "config.setSignupDisabled")
}

func (s *Server) handleConfigGetSSOSignupDisabled(w http.ResponseWriter, r *http.Request) {
	s.handleConfigBooleanFallback(w, r, "config.getSsoSignupDisabled", "DISABLE_SSO_SIGNUP", false)
}

func (s *Server) handleConfigSetSSOSignupDisabled(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "config.setSsoSignupDisabled")
}

func (s *Server) handleConfigGetBasicAuthDisabled(w http.ResponseWriter, r *http.Request) {
	s.handleConfigBooleanFallback(w, r, "config.getBasicAuthDisabled", "DISABLE_BASIC_AUTH", false)
}

func (s *Server) handleConfigSetBasicAuthDisabled(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "config.setBasicAuthDisabled")
}

func (s *Server) handleConfigGetAuthProviders(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "config.getAuthProviders", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "config.getAuthProviders",
			},
		})
		return
	}

	providers := []map[string]any{}
	if os.Getenv("OIDC_CLIENT_ID") != "" && os.Getenv("OIDC_CLIENT_SECRET") != "" && os.Getenv("OIDC_DISCOVERY_URL") != "" {
		providers = append(providers, map[string]any{
			"id":      "oidc",
			"name":    "OIDC",
			"enabled": true,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    providers,
		"bridge": map[string]any{
			"fallback":  "go-local-config",
			"procedure": "config.getAuthProviders",
			"reason":    "upstream unavailable; using local auth provider availability",
		},
	})
}

func (s *Server) handleConfigGetAlwaysVisibleTools(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "config.getAlwaysVisibleTools", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "config.getAlwaysVisibleTools",
			},
		})
		return
	}

	parsed, fallbackErr := s.readLocalMCPConfigObject()
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
			"detail":  fallbackErr.Error(),
		})
		return
	}

	settings, _ := parsed["settings"].(map[string]any)
	toolSelection, _ := settings["toolSelection"].(map[string]any)
	preferences := normalizeToolPreferences(toolSelection)
	alwaysVisible := normalizeAlwaysLoadedTools(preferences["alwaysLoadedTools"])
	if len(alwaysVisible) == 0 {
		alwaysVisible = normalizeToolNameList(parsed["alwaysVisibleTools"])
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    alwaysVisible,
		"bridge": map[string]any{
			"fallback":  "go-local-jsonc",
			"procedure": "config.getAlwaysVisibleTools",
			"reason":    "upstream unavailable; using local JSONC always-visible tool preferences",
		},
	})
}

func (s *Server) handleConfigSetAlwaysVisibleTools(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "config.setAlwaysVisibleTools")
}

func (s *Server) handleConfigBooleanFallback(w http.ResponseWriter, r *http.Request, procedure, key string, defaultValue bool) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), procedure, nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge":  map[string]any{"upstreamBase": upstreamBase, "procedure": procedure},
		})
		return
	}

	value, fallbackErr := s.localConfigBool(key, defaultValue)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    value,
		"bridge":  map[string]any{"fallback": "go-local-config-db", "procedure": procedure, "reason": "upstream unavailable; using local tormentnexus config value"},
	})
}

func (s *Server) handleConfigScalarFallback(w http.ResponseWriter, r *http.Request, procedure, key string, defaultValue int) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), procedure, nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge":  map[string]any{"upstreamBase": upstreamBase, "procedure": procedure},
		})
		return
	}

	value, fallbackErr := s.localConfigInt(key, defaultValue)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    value,
		"bridge":  map[string]any{"fallback": "go-local-config-db", "procedure": procedure, "reason": "upstream unavailable; using local tormentnexus config value"},
	})
}

func (s *Server) localConfigList() ([]map[string]any, error) {
	db, err := sql.Open("sqlite", s.localTormentNexusDBPath())
	if err != nil {
		return nil, err
	}
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA busy_timeout=5000")
	defer db.Close()

	rows, err := db.Query(`SELECT id, value FROM config ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []map[string]any{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"key": key, "value": value})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Server) localConfigValue(key string) (any, error) {
	db, err := sql.Open("sqlite", s.localTormentNexusDBPath())
	if err != nil {
		return nil, err
	}
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA busy_timeout=5000")
	defer db.Close()

	row := db.QueryRow(`SELECT value FROM config WHERE id = ? LIMIT 1`, key)
	var value string
	if err := row.Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return value, nil
}

func (s *Server) localConfigBool(key string, defaultValue bool) (bool, error) {
	value, err := s.localConfigValue(key)
	if err != nil || value == nil {
		return defaultValue, err
	}
	return value.(string) == "true", nil
}

func (s *Server) localConfigInt(key string, defaultValue int) (int, error) {
	value, err := s.localConfigValue(key)
	if err != nil || value == nil {
		return defaultValue, err
	}
	parsed, parseErr := strconv.Atoi(value.(string))
	if parseErr != nil {
		return defaultValue, nil
	}
	return parsed, nil
}
