package client

import (
	"strings"
	"testing"
)

func TestParseEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   string
		wantTarget string
		wantTLS    bool
	}{
		{name: "plain host", endpoint: "localhost:3478", wantTarget: "localhost:3478"},
		{name: "https scheme", endpoint: "https://api.example.com:443", wantTarget: "api.example.com:443", wantTLS: true},
		{name: "grpc scheme", endpoint: "grpc://127.0.0.1:3478", wantTarget: "127.0.0.1:3478"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed := parseEndpoint(test.endpoint)
			if parsed.Target != test.wantTarget || parsed.UseTLS != test.wantTLS {
				t.Fatalf("parseEndpoint(%q) = %#v, want target=%q tls=%t", test.endpoint, parsed, test.wantTarget, test.wantTLS)
			}
		})
	}
}

func TestTrimBearerPrefix(t *testing.T) {
	tests := []struct {
		token string
		want  string
	}{
		{token: "", want: ""},
		{token: "token", want: "token"},
		{token: "Bearer abc", want: "abc"},
		{token: "bearer   abc ", want: "abc"},
	}

	for _, test := range tests {
		if got := trimBearerPrefix(test.token); got != test.want {
			t.Fatalf("trimBearerPrefix(%q) = %q, want %q", test.token, got, test.want)
		}
	}
}

func TestDialOptionsRequireBothCertFiles(t *testing.T) {
	_, err := dialOptions(Params{
		Endpoint: "localhost:3478",
		CertPath: "/tmp/client.crt",
	})
	if err == nil || !strings.Contains(err.Error(), "both cert_path and cert_key_path") {
		t.Fatalf("dialOptions() error = %v, want missing key error", err)
	}
}

func TestDialOptionsRequireTLSForToken(t *testing.T) {
	_, err := dialOptions(Params{
		Endpoint: "localhost:3478",
		Token:    "secret",
	})
	if err == nil || !strings.Contains(err.Error(), "bearer token requires TLS") {
		t.Fatalf("dialOptions() error = %v, want TLS error", err)
	}
}

func TestDialOptionsAllowTokenWithTLSScheme(t *testing.T) {
	options, err := dialOptions(Params{
		Endpoint: "https://permify.example:3478",
		Token:    "secret",
	})
	if err != nil {
		t.Fatalf("dialOptions() error = %v", err)
	}
	if len(options) < 2 {
		t.Fatalf("dialOptions() returned %d options, want at least 2", len(options))
	}
}
