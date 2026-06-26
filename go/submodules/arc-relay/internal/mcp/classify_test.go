package mcp

import "testing"

func TestClassifyEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		endpointType string
		epName       string
		description  string
		want         string
	}{
		// Resources and prompts are always "read"
		{"resource always read", "resource", "config://main", "", "read"},
		{"prompt always read", "prompt", "deploy_prompt", "", "read"},

		// Tool read keywords
		{"tool get_users", "tool", "get_users", "", "read"},
		{"tool list_items", "tool", "list_items", "", "read"},
		{"tool search_logs", "tool", "search_logs", "", "read"},
		{"tool fetch_data", "tool", "fetch_data", "", "read"},
		{"tool query_db", "tool", "query_db", "", "read"},
		{"tool show_status", "tool", "show_status", "", "read"},
		{"tool describe_table", "tool", "describe_table", "", "read"},
		{"tool check_health", "tool", "check_health", "", "read"},
		{"tool count_records", "tool", "count_records", "", "read"},
		{"tool view_log", "tool", "view_log", "", "read"},
		{"tool lookup_user", "tool", "lookup_user", "", "read"},

		// Tool admin keywords
		{"tool delete_user", "tool", "delete_user", "", "admin"},
		{"tool remove_item", "tool", "remove_item", "", "admin"},
		{"tool destroy_session", "tool", "destroy_session", "", "admin"},
		{"tool execute_command", "tool", "execute_command", "", "admin"},
		{"tool reboot_server", "tool", "reboot_server", "", "admin"},
		{"tool shutdown_host", "tool", "shutdown_host", "", "admin"},
		{"tool restart_service", "tool", "restart_service", "", "admin"},
		{"tool kill_process", "tool", "kill_process", "", "admin"},
		{"tool deploy_app", "tool", "deploy_app", "", "admin"},

		// Default to write
		{"tool create_user", "tool", "create_user", "", "write"},
		{"tool update_config", "tool", "update_config", "", "write"},
		{"tool send_message", "tool", "send_message", "", "write"},

		// Description-based fallbacks
		{"description retrieves", "tool", "do_something", "Retrieves user data", "read"},
		{"description returns", "tool", "do_something", "Returns the list of items", "read"},
		{"description deletes", "tool", "do_something", "Deletes the resource permanently", "admin"},
		{"description executes", "tool", "do_something", "Executes a shell command", "admin"},

		// Case-insensitive matching
		{"uppercase tool name", "tool", "GET_USERS", "", "read"},
		{"mixed case tool name", "tool", "Delete_User", "", "admin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyEndpoint(tt.endpointType, tt.epName, tt.description)
			if got != tt.want {
				t.Errorf("ClassifyEndpoint(%q, %q, %q) = %q, want %q",
					tt.endpointType, tt.epName, tt.description, got, tt.want)
			}
		})
	}
}

func TestContainsWord(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		keyword string
		want    bool
	}{
		// Exact match
		{"exact match", "get", "get", true},
		{"exact match list", "list", "list", true},

		// Prefix match (keyword at start, followed by separator)
		{"prefix with underscore", "get_users", "get", true},
		{"prefix with hyphen", "get-users", "get", true},

		// Suffix match (keyword at end, preceded by separator)
		{"suffix with underscore", "user_list", "list", true},
		{"suffix with hyphen", "user-list", "list", true},

		// Middle segment
		{"middle segment", "my_get_users", "get", true},
		{"middle segment hyphen", "my-get-users", "get", true},

		// Should NOT match (no word boundary)
		{"no boundary getaway", "getaway", "get", false},
		{"no boundary target", "target", "get", false},
		{"no boundary forget", "forget", "get", false},
		{"no boundary listing", "listing", "list", false},
		{"no boundary blacklist", "blacklist", "list", false},

		// Empty cases
		{"empty name", "", "get", false},
		{"keyword not present", "create_user", "get", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsWord(tt.input, tt.keyword)
			if got != tt.want {
				t.Errorf("containsWord(%q, %q) = %v, want %v",
					tt.input, tt.keyword, got, tt.want)
			}
		})
	}
}
