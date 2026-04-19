package client

import "testing"

func TestNormalizeGRPCEndpoint(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"localhost:3478", "localhost:3478"},
		{"https://api.example.com:443", "api.example.com:443"},
		{"  grpc://127.0.0.1:9999  ", "127.0.0.1:9999"},
		{"HTTP://HOST:1", "HOST:1"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := normalizeGRPCEndpoint(tc.in); got != tc.want {
			t.Errorf("normalizeGRPCEndpoint(%q) = %q, want %q", tc.in, got, tc.want)
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
