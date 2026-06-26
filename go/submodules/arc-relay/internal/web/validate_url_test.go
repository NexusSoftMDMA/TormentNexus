package web

import (
	"net"
	"strings"
	"testing"
)

func TestValidateExternalURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		// Scheme checks (no DNS needed)
		{name: "http rejected", url: "http://example.com", wantErr: true, errMsg: "only https"},
		{name: "ftp rejected", url: "ftp://example.com", wantErr: true, errMsg: "only https"},
		{name: "no scheme", url: "example.com", wantErr: true, errMsg: "only https"},
		{name: "empty string", url: "", wantErr: true, errMsg: "only https"},
		{name: "no host", url: "https://", wantErr: true, errMsg: "invalid URL"},

		// Literal private IPs (no DNS needed)
		{name: "loopback IPv4", url: "https://127.0.0.1/path", wantErr: true, errMsg: "private/loopback"},
		{name: "loopback IPv6", url: "https://[::1]/path", wantErr: true, errMsg: "private/loopback"},
		{name: "private 10.x", url: "https://10.0.0.1/path", wantErr: true, errMsg: "private/loopback"},
		{name: "private 172.16.x", url: "https://172.16.0.1/path", wantErr: true, errMsg: "private/loopback"},
		{name: "private 192.168.x", url: "https://192.168.1.1/path", wantErr: true, errMsg: "private/loopback"},
		{name: "link-local", url: "https://169.254.169.254/latest", wantErr: true, errMsg: "private/loopback"},
		{name: "unspecified", url: "https://0.0.0.0/path", wantErr: true, errMsg: "private/loopback"},

		// Literal public IP (no DNS needed)
		{name: "public IP allowed", url: "https://8.8.8.8/path", wantErr: false},

		// Hostname resolving to loopback
		{name: "localhost hostname", url: "https://localhost/path", wantErr: true},

		// Unresolvable hostname
		{name: "unresolvable host", url: "https://this-host-does-not-exist-xyzzy.invalid/path", wantErr: true, errMsg: "cannot resolve"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExternalURL(tt.url)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for %q, got nil", tt.url)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for %q: %v", tt.url, err)
			}
			if tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip      string
		private bool
	}{
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"169.254.169.254", true},
		{"0.0.0.0", true},
		{"::1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("could not parse IP %q", tt.ip)
			}
			got := isPrivateIP(ip)
			if got != tt.private {
				t.Errorf("isPrivateIP(%s) = %v, want %v", tt.ip, got, tt.private)
			}
		})
	}
}
