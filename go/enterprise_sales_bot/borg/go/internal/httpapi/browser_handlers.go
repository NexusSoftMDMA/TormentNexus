package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleBrowserStatus(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "browser.status", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "browser.status",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Browser runtime is unavailable: upstream browser service is not available locally.",
		"data": map[string]any{
			"available": false,
			"active":    false,
			"pageCount": 0,
			"pageIds":   []string{},
		},
		"bridge": map[string]any{
			"fallback":  "go-local-browser",
			"procedure": "browser.status",
			"reason":    "upstream unavailable; browser service is not available locally",
		},
	})
}

func (s *Server) handleBrowserClosePage(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "browser.closePage")
}

func (s *Server) handleBrowserCloseAll(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "browser.closeAll", nil)
}

func (s *Server) handleBrowserSearchHistory(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing query parameter"})
		return
	}

	payload := map[string]any{"query": query}
	if maxResults := strings.TrimSpace(r.URL.Query().Get("maxResults")); maxResults != "" {
		parsed, err := strconv.Atoi(maxResults)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid maxResults query parameter"})
			return
		}
		payload["maxResults"] = parsed
	}
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "browser.searchHistory", payload)
}

func (s *Server) handleBrowserScrapePage(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "browser.scrapePage", nil)
}

func (s *Server) handleBrowserScreenshot(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "browser.screenshot", nil)
}

func (s *Server) handleBrowserDebug(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "browser.debug")
}

func (s *Server) handleBrowserProxyFetch(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "browser.proxyFetch")
}
