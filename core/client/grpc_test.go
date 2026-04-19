package client

import (
	"strings"
	"testing"
)

func TestParseGRPCEndpoint(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in            string
		wantTarget    string
		wantServerTLS bool
	}{
		{"localhost:3478", "localhost:3478", false},
		{"https://api.example.com:443", "api.example.com:443", true},
		{"grpcs://h:1", "h:1", true},
		{"  grpc://127.0.0.1:9999  ", "127.0.0.1:9999", false},
		{"HTTP://HOST:1", "HOST:1", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got := parseGRPCEndpoint(tc.in)
		if got.Target != tc.wantTarget || got.ServerTLSScheme != tc.wantServerTLS {
			t.Errorf("parseGRPCEndpoint(%q) = %#v, want target %q serverTLS %v", tc.in, got, tc.wantTarget, tc.wantServerTLS)
		}
	}
}

func TestStripBearerPrefix(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"  abc  ", "abc"},
		{"Bearer tok", "tok"},
		{"bearer  tok2 ", "tok2"},
		{"BEARER x", "x"},
		{"notbearer x", "notbearer x"},
	}
	for _, tc := range cases {
		if got := stripBearerPrefix(tc.in); got != tc.want {
			t.Errorf("stripBearerPrefix(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestDialOptionsPartialCertPaths(t *testing.T) {
	t.Parallel()
	_, _, err := dialOptions(Params{
		Endpoint:    "localhost:1",
		CertPath:    "/only/cert.pem",
		CertKeyPath: "",
	})
	if err == nil {
		t.Fatal("expected error when only cert path is set")
	}
}

func TestDialOptionsTokenRequiresTLS(t *testing.T) {
	t.Parallel()
	_, _, err := dialOptions(Params{
		Endpoint: "localhost:3478",
		Token:    "sekrit",
	})
	if err == nil || !strings.Contains(err.Error(), "bearer token requires TLS") {
		t.Fatalf("expected TLS requirement error, got %v", err)
	}
}

func TestDialOptionsTokenWithHTTPSScheme(t *testing.T) {
	t.Parallel()
	opts, tlsInUse, err := dialOptions(Params{
		Endpoint: "https://api.example.com:443",
		Token:    "tok",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !tlsInUse {
		t.Fatal("expected TLS in use")
	}
	if len(opts) < 2 {
		t.Fatalf("expected transport + PerRPC options, got %d", len(opts))
	}
}
